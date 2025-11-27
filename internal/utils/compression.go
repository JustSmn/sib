// Package utils предоставляет утилиты для сжатия данных с использованием Zstd.
// Zstd обеспечивает высокую скорость сжатия и хорошее соотношение сжатия.
package utils

import (
	"github.com/klauspost/compress/zstd"
)

// Глобальные encoder и decoder для избежания повторного создания
var (
	encoder, _ = zstd.NewWriter(nil)
	decoder, _ = zstd.NewReader(nil)
)

// CompressZstd сжимает данные с помощью алгоритма Zstd.
// Возвращает сжатые данные или ошибку, если сжатие не удалось.
func CompressZstd(data []byte) ([]byte, error) {
	return encoder.EncodeAll(data, make([]byte, 0, len(data))), nil
}

// DecompressZstd декомпрессирует данные, сжатые алгоритмом Zstd.
// Возвращает оригинальные данные или ошибку, если декомпрессия не удалась.
func DecompressZstd(data []byte) ([]byte, error) {
	return decoder.DecodeAll(data, nil)
}

// GetCompressionRatio вычисляет коэффициент сжатия.
// Возвращает отношение размера сжатых данных к оригинальным (0.0 - 1.0).
func GetCompressionRatio(original, compressed []byte) float64 {
	if len(original) == 0 {
		return 0
	}
	return float64(len(compressed)) / float64(len(original))
}
