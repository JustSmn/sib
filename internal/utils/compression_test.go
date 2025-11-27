package utils

import (
	"testing"
)

func TestCompressDecompressZstd(t *testing.T) {
	original := []byte("this is a test string for compression")

	compressed, err := CompressZstd(original)
	if err != nil {
		t.Fatalf("CompressZstd failed: %v", err)
	}

	if len(compressed) == 0 {
		t.Error("Compressed data is empty")
	}

	decompressed, err := DecompressZstd(compressed)
	if err != nil {
		t.Fatalf("DecompressZstd failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Decompressed data doesn't match original")
	}
}

func TestGetCompressionRatio(t *testing.T) {
	original := []byte("test data")
	compressed := []byte("compressed")

	ratio := GetCompressionRatio(original, compressed)

	if ratio <= 0 {
		t.Error("Compression ratio should be positive")
	}
}

func TestEmptyDataCompression(t *testing.T) {
	empty := []byte{}

	compressed, err := CompressZstd(empty)
	if err != nil {
		t.Fatalf("CompressZstd failed for empty data: %v", err)
	}

	decompressed, err := DecompressZstd(compressed)
	if err != nil {
		t.Fatalf("DecompressZstd failed for empty data: %v", err)
	}

	if len(decompressed) != 0 {
		t.Error("Decompressed empty data should be empty")
	}
}
