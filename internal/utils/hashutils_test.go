package utils

import (
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	data := []byte("hello world")
	hash := CalculateSHA256(data)

	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("Expected %s, got %s", expected, hash)
	}
}

func TestValidateHash(t *testing.T) {
	data := []byte("test data")
	validHash := CalculateSHA256(data)

	if !ValidateHash(data, validHash) {
		t.Error("ValidateHash should return true for valid hash")
	}

	if ValidateHash(data, "invalidhash") {
		t.Error("ValidateHash should return false for invalid hash")
	}
}
