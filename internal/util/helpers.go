package util

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func CreateHashedKey(text string) string {
	hash := sha256.Sum256([]byte(text))
	return "chat_history:" + hex.EncodeToString(hash[:])
}

func FormatDate(t time.Time) string {
	loc, _ := time.LoadLocation("Africa/Lagos")
	return t.In(loc).Format("January 02, 2006")
}
