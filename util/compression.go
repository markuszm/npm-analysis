package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strings"
)

func Compress(value string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(value)); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	data := b.String()
	return data, nil
}

func Decompress(value string) (string, error) {
	r := strings.NewReader(value)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}

	s, err := ioutil.ReadAll(gz)
	if err != nil {
		return "", err
	}

	if err := gz.Close(); err != nil {
		return "", err
	}

	data := string(s)
	return data, nil
}
