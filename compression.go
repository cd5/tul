package main

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	cdata := buf.Bytes()
	return cdata, nil
}

func decompress(cdata []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(cdata))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}
	data := buf.Bytes()
	return data, nil
}
