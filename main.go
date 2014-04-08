package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	AWS       = "ec2-54-72-96-37.eu-west-1.compute.amazonaws.com:4000"
	LOCALHOST = ":4000"
)

var serverAddress = AWS

func main() {
	log.SetFlags(0)

	err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := tls.Config{InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", serverAddress, &tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	c := &Conn{
		conn: conn,
		r:    bufio.NewReader(conn),
		buf:  make([]byte, 1<<12),
	}

	c.writeCmd("protocol version")
	c.writeString("0")
	c.writeString(config.email)
	c.writeString(config.password)
	c.readResponse()

	if len(os.Args) >= 2 {
		c.compile(os.Args[1:])
	}
	c.Close()
}

type Conn struct {
	conn net.Conn
	r    io.Reader
	buf  []byte
}

func (c *Conn) uploadFile(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	c.writeCmd("upload file")
	c.writeString(absPath)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	data, err = compress(data)
	if err != nil {
		log.Fatal(err)
	}
	c.writeInt32(len(data))
	c.write(data)
	c.readResponse()
}

func (c *Conn) compile(args []string) {
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	args1 := make([]string, 0, len(args))
	options := true
	for i := 0; i < len(args); {
		switch {
		case options && args[i] == "--":
			args1 = append(args1, args[i])
			options = false
			i++
		case options && strings.HasPrefix(args[i], "-"):
			{
				args1 = append(args1, args[i])
				switch {
				case args[i] == "-o":
					if !(i+1 < len(args)) {
						goto invalidOption
					}
					absPath, err := filepath.Abs(args[i+1])
					if err != nil {
						log.Fatal(err)
					}
					args1 = append(args1, absPath)
					i += 2
					continue
				case strings.HasPrefix(args[i], "-f"):
					i += 1
					continue
				default:
					goto invalidOption
				}
			invalidOption:
				log.Fatal("invalid command line option: " + args[i])
			}
		default:
			c.uploadFile(args[i])
			absPath, err := filepath.Abs(args[i])
			if err != nil {
				log.Fatal(err)
			}
			args1 = append(args1, absPath)
			i++
		}
	}

	c.writeCmd("compile")
	c.writeString(workdir)
	c.writeStrings(args1)
	c.readResponse()
}

func (c *Conn) Close() {
	c.writeCmd("close")

	c.readResponse()

	err := c.conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Conn) readResponse() string {
	for {
		resp, err := c.readString()
		if err != nil {
			log.Fatal(err)
		}

		switch resp {
		case "done":
			return resp
		case "network error", "server error":
			message, err := c.readString()
			if err != nil {
				log.Fatal(err)
			}
			log.Fatal(resp + ": " + message)
		case "write to stdout", "write to stderr":
			data, err := c.readString()
			if err != nil {
				log.Fatal(err)
			}
			var f *os.File
			if resp == "write to stdout" {
				f = os.Stdout
			} else {
				f = os.Stderr
			}
			f.Write([]byte(data))
		case "download file":
			path, err := c.readString()
			if err != nil {
				log.Fatal(err)
			}
			mode, err := c.readInt32()
			if err != nil {
				log.Fatal(err)
			}
			size, err := c.readInt32()
			if err != nil {
				log.Fatal(err)
			}
			data, err := c.readN(size)
			if err != nil {
				log.Fatal(err)
			}
			data, err = decompress(data)
			if err != nil {
				log.Fatal(err)
			}
			c.readResponse()

			err = ioutil.WriteFile(path, data, os.FileMode(mode&0777))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (c *Conn) readString() (string, error) {
	var buf []byte
	for {
		b, err := c.readByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			break
		} else {
			buf = append(buf, b)
		}
	}
	return string(buf), nil
}

func (c *Conn) readInt32() (int, error) {
	var n int
	for i := uint(0); i < 4; i++ {
		b, err := c.readByte()
		if err != nil {
			return 0, err
		}
		n |= int(b) << (8 * i)
	}
	return n, nil
}

func (c *Conn) readByte() (byte, error) {
	n, err := c.r.Read(c.buf[0:1])
	if n == 0 {
		if err == io.EOF {
			return 0, errors.New("network error: unexpected end of input stream")
		}
		return 0, errors.New("network error: read error")
	}
	if err != nil && err != io.EOF {
		return 0, err
	}
	return c.buf[0], nil
}

func (c *Conn) readN(N int) ([]byte, error) {
	buf := make([]byte, N)
	i := 0
	for i < N {
		n, err := c.r.Read(buf[i:])
		i += n
		if err == io.EOF {
			if i == N {
				return buf, nil
			}
			return nil, errors.New("network error: unexpected end of input stream")
		}
		if err != nil {
			return nil, errors.New("network error: read error")
		}
	}
	return buf, nil
}

func (c *Conn) writeCmd(cmd string) {
	c.writeString(cmd)
}

func (c *Conn) writeString(s string) {
	c.write([]byte(s))
	c.write([]byte{0})
}

func (c *Conn) writeStrings(s []string) {
	c.writeInt32(len(s))
	for i := range s {
		c.writeString(s[i])
	}
}

func (c *Conn) writeInt32(n int) {
	c.write([]byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)})
}

func (c *Conn) write(p []byte) {
	_, err := c.conn.Write(p)
	if err != nil {
		var buf [1 << 12]byte
		n, readErr := c.r.Read(buf[0:len("network error\000")])
		switch string(buf[0:n]) {
		case "network error\000":
			n, readErr = c.r.Read(buf[:])
			if n >= 1 && readErr == io.EOF || readErr == nil {
				log.Println("network error: " + string(buf[0:n-1]))
				os.Exit(1)
			}
		}
		log.Fatal(err)
	}
}
