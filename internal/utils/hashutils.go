// Package utils предоставляет утилиты для работы с хешированием SHA-256.
package utils

import (
	"crypto/sha256"
	"fmt"
)

// CalculateSHA256 вычисляет SHA-256 хеш от данных и возвращает в hex-формате.
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// ValidateHash проверяет, соответствуют ли данные указанному хешу.
// Возвращает true если хеш совпадает, false в противном случае.
func ValidateHash(data []byte, expectedHash string) bool {
	actualHash := CalculateSHA256(data)
	return actualHash == expectedHash
}
