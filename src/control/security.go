package main

import (
	"crypto/sha1"
	"encoding/base64"
)

func SHAify(input string) string {
	hasher := sha1.New()
	hasher.Write([]byte(input))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))[0:8]
}
