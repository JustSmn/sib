// Package utils предоставляет кросс-платформенные утилиты для работы с файловой системой,
// сжатием данных и хешированием. Все функции разработаны для работы на Windows, Linux и macOS.
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// CreateDirIfNotExists создает директорию со всеми родительскими директориями, если она не существует.
// Автоматически работает с путями для текущей ОС (Windows: `\`, Unix: `/`).
func CreateDirIfNotExists(dir string) error {
	if FileExists(dir) {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

// FileExists проверяет существование файла или директории.
// Использует os.Stat, который работает одинаково на всех ОС.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// WriteFileAtomic атомарно записывает файл через временный файл.
// Это защищает от частичной записи при сбоях. На Windows использует Rename,
// который атомарен в NTFS. На Unix системах rename() также атомарен.
func WriteFileAtomic(path string, data []byte) error {

	// Создаем временный файл в той же директории
	dir := filepath.Dir(path)

	tmpFile, err := os.CreateTemp(dir, "tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()

	// Записываем данные во временный файл
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Закрываем файл
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Атомарно переименовываем временный файл в целевой
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// ReadFile читает весь файл в память. Работает одинаково на всех ОС.
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// RemoveFile удаляет файл.
func RemoveFile(path string) error { return os.Remove(path) }

// ListFiles возвращает список файлов (не директорий) в указанной директории.
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// IsWindows проверяет, работает ли программа на Windows.
// Полезно для OS-specific логики, если понадобится.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// NormalizePath нормализует путь для текущей ОС.
// На Windows заменяет `/` на `\`, на Unix оставляет как есть.
func NormalizePath(path string) string {
	return filepath.Clean(path)
}
