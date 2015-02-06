package main

import (
	"crypto/sha1"
	"encoding/base64"

	"github.com/fcoury/rtc-go/encrypt"
)

func makeKey(key string) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(key))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return []byte(sha[0:16])
}

func Encrypt(key string, s string) string {
	secretKey := makeKey(key)
	encrypted, err := encrypt.Encrypt(secretKey, []byte(s))
	if err != nil {
		return s
	}
	res := encrypt.EncodeBase64(encrypted)
	// fmt.Println("Enc", res)
	return res
}

func Decrypt(key string, s string) string {
	secretKey := makeKey(key)
	// fmt.Println("Got", s)
	decoded, err := encrypt.DecodeBase64(s)
	if err != nil {
		return s
	}
	res, err := encrypt.Decrypt(secretKey, decoded)
	if err != nil {
		return s
	}
	// fmt.Println("Dec", res)
	return res
}
