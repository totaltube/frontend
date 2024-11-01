package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
)

func Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
func FromBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
func Base64Url(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}
func FromBase64Url(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}

func Flate(data []byte) []byte {
	var b bytes.Buffer
	zw, _ := flate.NewWriter(&b, flate.DefaultCompression)
	_, err := zw.Write(data)
	if err != nil {
		log.Println(err)
		return nil
	}
	_ = zw.Close()
	return b.Bytes()
}

func Bytes(data interface{}) []byte {
	return []byte(fmt.Sprintf("%v", data))
}

func Deflate(data []byte) []byte {
	var b = bytes.NewReader(data)
	zr := flate.NewReader(b)
	s, _ := ioutil.ReadAll(zr)
	return s
}

func Gzip(data []byte) []byte {
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	_, err := zw.Write(data)
	if err != nil {
		log.Println(err)
		return nil
	}
	_ = zw.Close()
	return b.Bytes()
}

func Ungzip(data []byte) []byte {
	var b = bytes.NewReader(data)
	zr, _ := gzip.NewReader(b)
	s, _ := ioutil.ReadAll(zr)
	return s
}

func Zip(data []byte) []byte {
	var b bytes.Buffer
	zw := zlib.NewWriter(&b)
	_, err := zw.Write(data)
	if err != nil {
		log.Println(err)
		return nil
	}
	_ = zw.Close()
	return b.Bytes()
}

func Unzip(data []byte) []byte {
	var b = bytes.NewReader(data)
	zr, _ := zlib.NewReader(b)
	s, _ := ioutil.ReadAll(zr)
	return s
}
