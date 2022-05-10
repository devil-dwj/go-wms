package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/spaolacci/murmur3"
)

func Hash(data []byte) uint64 {
	return murmur3.Sum64(data)
}

func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

func Md5Hex(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}

func HmacSha1(src string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(src))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
