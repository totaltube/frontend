package helpers

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
)

func Sha1Hash(s string) (hash string) {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func Md5Hash(s string) (hash string) {
	h := md5.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}
