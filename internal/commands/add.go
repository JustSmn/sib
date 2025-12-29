package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sib/internal/core/index"
	"sib/internal/core/objects"
	"sib/internal/core/storage"
)

func Add(repoPath string) error {
	if repoPath == "" {
		repoPath = "."
	}

	// Проверяем, что это sib репозиторий
	sibDir := filepath.Join(repoPath, ".sib")
	if _, err := os.Stat(sibDir); os.IsNotExist(err) {
		return fmt.Errorf("not a sib repository")
	}

	// Загружаем индекс
	idx, err := index.NewIndex(repoPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Создаем хранилище объектов
	store, err := storage.NewObjectStore(repoPath)
	if err != nil {
		return fmt.Errorf("failed to create object store: %w", err)
	}

	// Сканируем все файлы
	addedCount := 0
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем:
		// - Директории
		// - Файл .sib и всё внутри
		// - Скрытые файлы (начинающиеся с .)
		if info.IsDir() {
			if info.Name() == ".sib" {
				return filepath.SkipDir
			}
			return nil
		}

		// Пропускаем файлы внутри .sib
		if isInsideSibDir(path, repoPath) {
			return nil
		}

		// Пропускаем скрытые файлы (опционально)
		if filepath.Base(path)[0] == '.' && filepath.Base(path) != "." {
			return nil
		}

		// Получаем относительный путь
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil // пропускаем ошибки
		}

		// Читаем файл
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("warning: could not read %s: %v\n", relPath, err)
			return nil
		}

		// Создаем blob
		blob := objects.NewBlob(content)

		// Сохраняем в хранилище
		hash, err := store.WriteObject(blob)
		if err != nil {
			fmt.Printf("warning: could not save %s: %v\n", relPath, err)
			return nil
		}

		// Определяем режим файла
		mode := index.DetectFileMode(info)

		// Добавляем в индекс
		if err := idx.Add(relPath, hash.String(), info.Size(), mode, info.ModTime()); err != nil {
			fmt.Printf("warning: could not add %s to index: %v\n", relPath, err)
			return nil
		}

		addedCount++
		fmt.Printf("added %s\n", relPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Сохраняем индекс
	if err := idx.Save(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Printf("Added %d files to index\n", addedCount)
	return nil
}

func isInsideSibDir(path, repoPath string) bool {
	sibDir := filepath.Join(repoPath, ".sib")
	rel, err := filepath.Rel(sibDir, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
