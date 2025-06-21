package helpers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"log"
	"sync"

	"golang.org/x/crypto/md4" //nolint
	"sersh.com/totaltube/frontend/internal"
)

func Sha1Hash(s string) (hash string) {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func Sha1HashRaw(s string) (hash []byte) {
	h := sha1.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func Md5Hash(s string) (hash string) {
	h := md5.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}
func Md5HashRaw(s string) (hash []byte) {
	h := md5.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}
func Md4Hash(s string) (hash string) {
	h := md4.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}
func Md4HashRaw(s string) (hash []byte) {
	h := md4.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}
func Sha256Hash(s string) (hash string) {
	h := sha256.Sum256([]byte(s))
	hash = hex.EncodeToString(h[:])
	return
}
func Sha256HashRaw(s string) (hash []byte) {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

func Sha512Hash(s string) (hash string) {
	h := sha512.Sum512([]byte(s))
	hash = hex.EncodeToString(h[:])
	return
}
func Sha512HashRaw(s string) (hash []byte) {
	h := sha512.Sum512([]byte(s))
	return h[:]
}

var encryptionKey string
var encryptMutex sync.Mutex

func EncryptBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(EncryptDecrypt(input)))
}

func DecryptBase64(input string) string {
	out, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		log.Println("can't decrypt base64 string:", err)
		return EncryptDecrypt(input)
	}
	return EncryptDecrypt(string(out))
}

func EncryptDecrypt(input string) (output string) {
	encryptMutex.Lock()
	defer encryptMutex.Unlock()
	if encryptionKey == "" {
		// prepare encryptionKey as md5 sum of secret key
		passPhrase := internal.Config.Frontend.SecretKey
		if passPhrase == "" {
			passPhrase = internal.Config.General.ApiSecret
		}
		h := md5.New()
		h.Write([]byte(passPhrase))
		encryptionKey = hex.EncodeToString(h.Sum(nil))
	}
	ek := []byte(encryptionKey)
	kL := len(ek)
	ib := []byte(input)
	for i := range ib {
		output += string(ib[i] ^ ek[i%kL])
	}
	return output
}
