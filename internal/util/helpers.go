package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func CreateHashedKey(text string) string {
	hash := sha256.Sum256([]byte(text))
	return "chat_history:" + hex.EncodeToString(hash[:])
}
