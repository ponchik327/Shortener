package service

import (
	"crypto/rand"
	"fmt"
)

const (
	_base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	_base62Len      = len(_base62Alphabet)
)

// generateCode генерирует случайную строку в кодировке base62 заданной длины.
func generateCode(length int) (string, error) {
	result := make([]byte, length)

	if _, err := rand.Read(result); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	for i, b := range result {
		result[i] = _base62Alphabet[int(b)%_base62Len]
	}

	return string(result), nil
}
