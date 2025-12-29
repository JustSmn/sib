// index_test.go
package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewIndex(t *testing.T) {
	// Создаем временную директорию для тестов
	tmpDir := t.TempDir()

	t.Run("Create new index in empty directory", func(t *testing.T) {
		idx, err := NewIndex(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}

		// Проверяем, что .sib директория создана
		sibDir := filepath.Join(tmpDir, ".sib")
		if _, err := os.Stat(sibDir); os.IsNotExist(err) {
			t.Error(".sib directory was not created")
		}

		// Проверяем, что файл индекса создан
		indexPath := filepath.Join(sibDir, "index")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			t.Error("index file was not created")
		}

		// Проверяем начальное состояние
		if idx.Count() != 0 {
			t.Errorf("Expected 0 entries, got %d", idx.Count())
		}

		if idx.Path() != indexPath {
			t.Errorf("Expected path %s, got %s", indexPath, idx.Path())
		}
	})

	t.Run("Load existing index", func(t *testing.T) {
		// Сначала создаем индекс
		idx1, err := NewIndex(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create first index: %v", err)
		}

		// Добавляем запись
		err = idx1.Add("test.txt", "abc123", 1024, "100644", time.Now())
		if err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}

		// Сохраняем
		if err := idx1.Save(); err != nil {
			t.Fatalf("Failed to save index: %v", err)
		}

		// Загружаем заново
		idx2, err := NewIndex(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load existing index: %v", err)
		}

		// Проверяем, что запись сохранилась
		if idx2.Count() != 1 {
			t.Errorf("Expected 1 entry after reload, got %d", idx2.Count())
		}

		entry, err := idx2.Get("test.txt")
		if err != nil {
			t.Errorf("Failed to get entry: %v", err)
		}

		if entry.Hash != "abc123" {
			t.Errorf("Expected hash abc123, got %s", entry.Hash)
		}
	})
}

func TestIndexAdd(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	now := time.Now()

	tests := []struct {
		name    string
		path    string
		hash    string
		size    int64
		mode    string
		mtime   time.Time
		wantErr bool
	}{
		{
			name:  "Valid file",
			path:  "main.go",
			hash:  "sha256-abc123",
			size:  1024,
			mode:  "100644",
			mtime: now,
		},
		{
			name:  "Executable file",
			path:  "script.sh",
			hash:  "sha256-def456",
			size:  512,
			mode:  "100755",
			mtime: now,
		},
		{
			name:  "Directory",
			path:  "src",
			hash:  "sha256-ghi789",
			size:  0,
			mode:  "040000",
			mtime: now,
		},
		{
			name:    "Empty path",
			path:    "",
			hash:    "sha256-test",
			size:    100,
			mode:    "100644",
			mtime:   now,
			wantErr: true,
		},
		{
			name:    "Empty hash",
			path:    "file.txt",
			hash:    "",
			size:    100,
			mode:    "100644",
			mtime:   now,
			wantErr: true,
		},
		{
			name:    "Negative size",
			path:    "file.txt",
			hash:    "sha256-test",
			size:    -1,
			mode:    "100644",
			mtime:   now,
			wantErr: true,
		},
		{
			name:    "Invalid mode",
			path:    "file.txt",
			hash:    "sha256-test",
			size:    100,
			mode:    "999999",
			mtime:   now,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := idx.Add(tt.path, tt.hash, tt.size, tt.mode, tt.mtime)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Проверяем, что запись добавилась
			entry, err := idx.Get(tt.path)
			if err != nil {
				t.Errorf("Failed to get added entry: %v", err)
				return
			}

			if entry.Hash != tt.hash {
				t.Errorf("Hash mismatch: want %s, got %s", tt.hash, entry.Hash)
			}

			if entry.Size != tt.size {
				t.Errorf("Size mismatch: want %d, got %d", tt.size, entry.Size)
			}

			if entry.Mode != tt.mode {
				t.Errorf("Mode mismatch: want %s, got %s", tt.mode, entry.Mode)
			}
		})
	}

	t.Run("Update existing entry", func(t *testing.T) {
		// Добавляем первый раз
		err := idx.Add("update.txt", "hash1", 100, "100644", now)
		if err != nil {
			t.Fatalf("Failed to add first time: %v", err)
		}

		// Обновляем с новыми данными
		newTime := now.Add(time.Hour)
		err = idx.Add("update.txt", "hash2", 200, "100755", newTime)
		if err != nil {
			t.Fatalf("Failed to update: %v", err)
		}

		entry, err := idx.Get("update.txt")
		if err != nil {
			t.Fatalf("Failed to get updated entry: %v", err)
		}

		if entry.Hash != "hash2" {
			t.Errorf("Hash not updated: want hash2, got %s", entry.Hash)
		}

		if entry.Size != 200 {
			t.Errorf("Size not updated: want 200, got %d", entry.Size)
		}
	})
}

func TestIndexRemove(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Добавляем несколько записей
	files := []string{"a.txt", "b.txt", "c.txt"}
	for i, file := range files {
		err := idx.Add(file, "hash"+file, int64(i*100), "100644", time.Now())
		if err != nil {
			t.Fatalf("Failed to add %s: %v", file, err)
		}
	}

	if idx.Count() != 3 {
		t.Errorf("Expected 3 entries, got %d", idx.Count())
	}

	// Удаляем одну запись
	t.Run("Remove existing file", func(t *testing.T) {
		err := idx.Remove("b.txt")
		if err != nil {
			t.Errorf("Failed to remove: %v", err)
		}

		if idx.Count() != 2 {
			t.Errorf("Expected 2 entries after removal, got %d", idx.Count())
		}

		// Проверяем, что удаленного файла нет
		_, err = idx.Get("b.txt")
		if err == nil {
			t.Error("Expected error when getting removed file")
		}
	})

	t.Run("Remove non-existent file", func(t *testing.T) {
		err := idx.Remove("nonexistent.txt")
		if err == nil {
			t.Error("Expected error when removing non-existent file")
		}
	})

	t.Run("Remove with empty path", func(t *testing.T) {
		err := idx.Remove("")
		if err == nil {
			t.Error("Expected error when removing with empty path")
		}
	})
}

func TestIndexSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаем и заполняем индекс
	idx1, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	now := time.Now()
	testEntries := []struct {
		path, hash, mode string
		size             int64
	}{
		{"src/main.go", "hash1", "100644", 1500},
		{"src/util.go", "hash2", "100644", 800},
		{"bin/app", "hash3", "100755", 2500},
		{"docs/README.md", "hash4", "100644", 300},
	}

	for _, entry := range testEntries {
		err := idx1.Add(entry.path, entry.hash, entry.size, entry.mode, now)
		if err != nil {
			t.Fatalf("Failed to add %s: %v", entry.path, err)
		}
	}

	// Сохраняем
	if err := idx1.Save(); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Проверяем, что файл существует
	indexPath := filepath.Join(tmpDir, ".sib", "index")
	info, err := os.Stat(indexPath)
	if err != nil {
		t.Fatalf("Index file not found: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Index file is empty")
	}

	// Загружаем заново
	idx2, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Проверяем целостность данных
	if idx1.Count() != idx2.Count() {
		t.Errorf("Entry count mismatch: original %d, loaded %d",
			idx1.Count(), idx2.Count())
	}

	// Проверяем каждую запись
	for _, entry := range testEntries {
		loaded, err := idx2.Get(entry.path)
		if err != nil {
			t.Errorf("Failed to get %s: %v", entry.path, err)
			continue
		}

		if loaded.Hash != entry.hash {
			t.Errorf("Hash mismatch for %s: want %s, got %s",
				entry.path, entry.hash, loaded.Hash)
		}

		if loaded.Size != entry.size {
			t.Errorf("Size mismatch for %s: want %d, got %d",
				entry.path, entry.size, loaded.Size)
		}
	}
}

func TestIndexDiff(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Создаем файлы в рабочей директории
	workingFiles := []string{
		"existing.txt", // Будет в индексе и на диске
		"modified.txt", // Будет в индексе, но изменится на диске
		"deleted.txt",  // Будет только в индексе
	}

	// Добавляем в индекс
	for _, file := range workingFiles {
		fullPath := filepath.Join(tmpDir, file)

		// Создаем файл
		content := []byte("content for " + file)
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}

		// Добавляем в индекс
		info, _ := os.Stat(fullPath)
		err := idx.Add(file, "hash-"+file, info.Size(), "100644", info.ModTime())
		if err != nil {
			t.Fatalf("Failed to add to index: %v", err)
		}
	}

	// Модифицируем один файл на диске
	modifiedPath := filepath.Join(tmpDir, "modified.txt")
	newContent := []byte("modified content")
	if err := os.WriteFile(modifiedPath, newContent, 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Удаляем один файл с диска
	deletedPath := filepath.Join(tmpDir, "deleted.txt")
	if err := os.Remove(deletedPath); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Добавляем новый файл только на диск
	newFilePath := filepath.Join(tmpDir, "newfile.txt")
	if err := os.WriteFile(newFilePath, []byte("new file"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// Выполняем diff
	added, modified, deleted, err := idx.Diff(tmpDir)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Проверяем результаты
	if len(added) != 1 || added[0] != "newfile.txt" {
		t.Errorf("Added files mismatch: want [newfile.txt], got %v", added)
	}

	if len(modified) != 1 || modified[0] != "modified.txt" {
		t.Errorf("Modified files mismatch: want [modified.txt], got %v", modified)
	}

	if len(deleted) != 1 || deleted[0] != "deleted.txt" {
		t.Errorf("Deleted files mismatch: want [deleted.txt], got %v", deleted)
	}
}

func TestIndexValidate(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Создаем валидный файл
	validFile := filepath.Join(tmpDir, "valid.txt")
	content := []byte("valid content")
	if err := os.WriteFile(validFile, content, 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	info, _ := os.Stat(validFile)
	err = idx.Add("valid.txt", "hash-valid", info.Size(), "100644", info.ModTime())
	if err != nil {
		t.Fatalf("Failed to add valid file: %v", err)
	}

	// Создаем невалидный файл (размер не совпадает)
	invalidFile := filepath.Join(tmpDir, "invalid.txt")
	if err := os.WriteFile(invalidFile, []byte("short"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	info, _ = os.Stat(invalidFile)
	// Добавляем с неправильным размером
	err = idx.Add("invalid.txt", "hash-invalid", info.Size()+100, "100644", info.ModTime())
	if err != nil {
		t.Fatalf("Failed to add invalid file: %v", err)
	}

	// Проверяем валидацию
	invalidFiles, err := idx.Validate(tmpDir)

	// Должна быть ошибка
	if err == nil {
		t.Error("Expected validation error")
	}

	// Должен быть один невалидный файл
	if len(invalidFiles) != 1 || invalidFiles[0] != "invalid.txt" {
		t.Errorf("Invalid files mismatch: want [invalid.txt], got %v", invalidFiles)
	}

	// Проверяем метод GetInvalidFiles
	invalidFiles2 := idx.GetInvalidFiles(tmpDir)
	if len(invalidFiles2) != 1 || invalidFiles2[0] != "invalid.txt" {
		t.Errorf("GetInvalidFiles mismatch: want [invalid.txt], got %v", invalidFiles2)
	}
}

func TestIndexGetAllEntries(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Добавляем записи в случайном порядке
	files := []string{"z.txt", "a.txt", "m.txt", "b.txt"}
	for _, file := range files {
		err := idx.Add(file, "hash-"+file, 100, "100644", time.Now())
		if err != nil {
			t.Fatalf("Failed to add %s: %v", file, err)
		}
	}

	// Получаем все записи
	entries := idx.GetAllEntries()

	// Проверяем количество
	if len(entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entries))
	}

	// Проверяем сортировку (должны быть в алфавитном порядке)
	expectedOrder := []string{"a.txt", "b.txt", "m.txt", "z.txt"}
	for i, entry := range entries {
		if entry.Path != expectedOrder[i] {
			t.Errorf("Entry %d: expected %s, got %s", i, expectedOrder[i], entry.Path)
		}
	}
}

func TestIndexEdgeCases(t *testing.T) {
	t.Run("Path normalization", func(t *testing.T) {
		tmpDir := t.TempDir()
		idx, err := NewIndex(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}

		// Добавляем с разными форматами путей
		testPaths := []string{
			"dir/file.txt",
			"dir\\file.txt",         // Windows-style
			"./dir/../dir/file.txt", // С точками
			"dir//file.txt",         // Двойной слэш
		}

		for _, path := range testPaths {
			err := idx.Add(path, "hash", 100, "100644", time.Now())
			if err != nil {
				t.Errorf("Failed to add path %s: %v", path, err)
			}
		}

		// Все пути должны нормализоваться к одному
		if idx.Count() != 1 {
			t.Errorf("Expected 1 unique entry after normalization, got %d", idx.Count())
		}
	})

	t.Run("Clear index", func(t *testing.T) {
		tmpDir := t.TempDir()
		idx, err := NewIndex(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}

		// Добавляем записи
		for i := 0; i < 5; i++ {
			err := idx.Add("file"+string(rune('a'+i))+".txt", "hash", 100, "100644", time.Now())
			if err != nil {
				t.Fatalf("Failed to add file: %v", err)
			}
		}

		if idx.Count() != 5 {
			t.Errorf("Expected 5 entries, got %d", idx.Count())
		}

		// Очищаем
		if err := idx.Clear(); err != nil {
			t.Errorf("Failed to clear: %v", err)
		}

		if idx.Count() != 0 {
			t.Errorf("Expected 0 entries after clear, got %d", idx.Count())
		}

		// Проверяем, что HasChanges работает
		if idx.HasChanges() {
			t.Error("HasChanges should return false after clear")
		}
	})
}

func TestIndexCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаем битый JSON файл
	indexPath := filepath.Join(tmpDir, ".sib", "index")
	sibDir := filepath.Dir(indexPath)

	if err := os.MkdirAll(sibDir, 0755); err != nil {
		t.Fatalf("Failed to create .sib directory: %v", err)
	}

	// Пишем битый JSON
	corruptJSON := []byte(`{"version": 1, "entries": { "test": { `)
	if err := os.WriteFile(indexPath, corruptJSON, 0644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Пытаемся загрузить
	idx, err := NewIndex(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load corrupt index: %v", err)
	}

	// Должен создать новый пустой индекс
	if idx.Count() != 0 {
		t.Errorf("Corrupt index should be replaced with empty, got %d entries", idx.Count())
	}
}
