package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func Init(repoPath string) error {
	if repoPath == "" {
		repoPath = "."
	}

	// Абсолютный путь для сообщений
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Проверяем, что это директория
	if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", absPath)
	}

	sibDir := filepath.Join(absPath, ".sib")

	// Проверяем, не инициализирован ли уже
	if _, err := os.Stat(sibDir); err == nil {
		return fmt.Errorf("already a sib repository")
	}

	// Создаем обязательные директории
	dirs := []string{
		filepath.Join(sibDir, "objects"),
		filepath.Join(sibDir, "refs", "heads"),
		filepath.Join(sibDir, "refs", "tags"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}
	}

	// Создаем HEAD файл
	headPath := filepath.Join(sibDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/master\n"), 0644); err != nil {
		return fmt.Errorf("failed to create HEAD: %w", err)
	}

	// Создаем базовый конфиг (опционально, можно пропустить)
	configPath := filepath.Join(sibDir, "config")
	configContent := "[core]\n\trepositoryformatversion = 0\n"
	_ = os.WriteFile(configPath, []byte(configContent), 0644) // Игнорируем ошибку

	fmt.Printf("Initialized empty Sib repository in %s\n", sibDir)
	return nil
}
