package helpers

import (
	"bytes"
	"encoding/base64"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"

	"io/ioutil"
	"log"
)

func Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Flate(data string) []byte {
	var b bytes.Buffer
	zw, _ := flate.NewWriter(&b, flate.DefaultCompression)
	_, err := zw.Write([]byte(data))
	if err != nil {
		log.Println(err)
		return nil
	}
	_ = zw.Close()
	return b.Bytes()
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