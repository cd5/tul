package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	pathutil "path"
	"strings"
)

type Config struct {
	email    string
	password string
}

var config Config

func readConfig() error {
	home := os.Getenv("HOME")
	if home == "" {
		return errors.New("missing HOME environment variable")
	}

	configFile := pathutil.Join(home, ".config", "tul", "account")
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		i := strings.IndexByte(line, '=')
		if i == -1 {
			return errors.New(fmt.Sprintf("%s:%d: syntax error", configFile, lineNumber+1))
		}
		key := line[:i]
		value := line[i+1:]

		switch key {
		case "email":
			config.email = value
		case "password":
			config.password = value
		default:
			return errors.New(fmt.Sprintf("%s:%d: invalid key: %s", configFile, lineNumber+1, key))
		}
	}

	if config.email == "" {
		return errors.New(fmt.Sprintf("%s: missing line: email=<your e-mail address>", configFile))
	}
	if config.password == "" {
		return errors.New(fmt.Sprintf("%s: missing line: password=<a password>", configFile))
	}

	return nil
}
