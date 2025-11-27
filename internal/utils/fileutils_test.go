package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir", "subdir")

	err := CreateDirIfNotExists(testDir)
	if err != nil {
		t.Fatalf("CreateDirIfNotExists failed: %v", err)
	}

	if !FileExists(testDir) {
		t.Error("Directory was not created")
	}
}

func TestFileExists(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "testfile.txt")

	// Файл не должен существовать
	if FileExists(tmpFile) {
		t.Error("FileExists should return false for non-existent file")
	}

	// Создаем файл
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !FileExists(tmpFile) {
		t.Error("FileExists should return true for existing file")
	}
}

func TestWriteFileAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	data := []byte("test data")

	err := WriteFileAtomic(testFile, data)
	if err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	// Проверяем, что файл создан и содержит правильные данные
	content, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(content))
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	expected := "hello world"

	if err := os.WriteFile(testFile, []byte(expected), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	content, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != expected {
		t.Errorf("Expected %s, got %s", expected, string(content))
	}
}

func TestRemoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := RemoveFile(testFile); err != nil {
		t.Fatalf("RemoveFile failed: %v", err)
	}

	if FileExists(testFile) {
		t.Error("File still exists after removal")
	}
}

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаем тестовые файлы
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Создаем поддиректорию (не должна попасть в результат)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	list, err := ListFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(list) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(list))
	}
}
