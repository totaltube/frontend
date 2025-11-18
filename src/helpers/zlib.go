package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

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
func Base64RawUrl(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
func FromBase64RawUrl(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
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

func Bytes(data any) []byte {
	return fmt.Appendf(nil, "%v", data)
}

func Deflate(data []byte) []byte {
	var b = bytes.NewReader(data)
	zr := flate.NewReader(b)
	s, _ := io.ReadAll(zr)
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
	s, _ := io.ReadAll(zr)
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
	s, _ := io.ReadAll(zr)
	return s
}

// HtmlEntitiesAll converts every character in the string to numeric HTML entities (&#code;)
func HtmlEntitiesAll(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteString("&#")
		b.WriteString(strconv.Itoa(int(r)))
		b.WriteString(";")
	}
	return b.String()
}
