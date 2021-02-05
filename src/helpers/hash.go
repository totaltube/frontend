package helpers

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"log"
	"sersh.com/totaltube/frontend/internal"
	"sync"
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
