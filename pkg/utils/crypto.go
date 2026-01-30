package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

// GenerateSalt creates a random string to be used as salt
func GenerateSalt() string {
	rand.Seed(time.Now().UnixNano())
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// HashPassword combines password and salt and returns MD5 hash
func HashPassword(password, salt string) string {
	hasher := md5.New()
	hasher.Write([]byte(password + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}
