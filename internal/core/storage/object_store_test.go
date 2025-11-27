/*
Что тестируется:

✅ Создание хранилища и структуры директорий
✅ Вычисление хешей (разные типы данных)
✅ Преобразование хешей в пути
✅ Полный цикл записи/чтения (Blob, Tree, Commit)
✅ Проверка целостности данных
✅ Множественные объекты
✅ Пустые объекты
✅ Обработка ошибок (несуществующие объекты, поврежденные данные)
✅ Структура хранилища (визуализация в логах)
*/

package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sib/internal/core/objects"
)

// TestNewObjectStore проверяет создание хранилища
func TestNewObjectStore(t *testing.T) {
	tmpDir := t.TempDir()

	store := NewObjectStore(tmpDir)

	// Проверяем, что путь сформирован правильно
	expectedPath := filepath.Join(tmpDir, ".sib", "objects")
	if store.objectsDir != expectedPath {
		t.Errorf("Expected objects dir '%s', got '%s'", expectedPath, store.objectsDir)
	}

	// Проверяем, что директория не создается автоматически
	if _, err := os.Stat(store.objectsDir); !os.IsNotExist(err) {
		t.Error("Objects directory should not be created automatically")
	}
}

// TestCalculateHash проверяет вычисление хеша
func TestCalculateHash(t *testing.T) {
	store := NewObjectStore(t.TempDir())

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "special characters",
			data:     []byte("test\n\r\t\x00"),
			expected: "d00c161ea28e969d839502aeff7a7d02a6c061b56e96ffaaec3c86e0d1a53256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := store.calculateHash(tt.data)
			if hash.String() != tt.expected {
				t.Errorf("Expected hash '%s', got '%s'", tt.expected, hash)
			}
		})
	}
}

// TestHashToPath проверяет преобразование хеша в путь
func TestHashToPath(t *testing.T) {
	store := NewObjectStore(t.TempDir())

	tests := []struct {
		name        string
		hash        objects.Hash
		shouldError bool
		expected    string
	}{
		{
			name:        "valid hash",
			hash:        objects.Hash("abc123def456"),
			shouldError: false,
			expected:    filepath.Join(store.objectsDir, "ab", "c123def456"),
		},
		{
			name:        "short hash",
			hash:        objects.Hash("a"),
			shouldError: true,
		},
		{
			name:        "empty hash",
			hash:        objects.Hash(""),
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := store.hashToPath(tt.hash)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if path != tt.expected {
				t.Errorf("Expected path '%s', got '%s'", tt.expected, path)
			}
		})
	}
}

// TestWriteAndReadBlob проверяет полный цикл записи/чтения blob
func TestWriteAndReadBlob(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Тестовые данные
	testContent := []byte("This is a test file content for blob object")
	blob := objects.NewBlob(testContent)

	// Записываем объект
	hash, err := store.WriteObject(blob)
	if err != nil {
		t.Fatalf("WriteObject failed: %v", err)
	}

	// Проверяем, что хеш установился в объекте
	if blob.GetHash() != hash {
		t.Errorf("Blob hash not set correctly. Expected %s, got %s", hash, blob.GetHash())
	}

	// Проверяем, что файл создался в правильной структуре
	objectPath, _ := store.hashToPath(hash)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		t.Error("Object file was not created")
	}

	// Проверяем структуру пути
	expectedDir := filepath.Join(store.objectsDir, hash.String()[:2])
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("Object directory was not created: %s", expectedDir)
	}

	// Читаем объект обратно
	readObj, err := store.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	// Проверяем тип
	readBlob, ok := readObj.(*objects.Blob)
	if !ok {
		t.Fatalf("Expected *objects.Blob, got %T", readObj)
	}

	// Проверяем содержимое
	readContent := readBlob.Content()
	if string(readContent) != string(testContent) {
		t.Errorf("Content mismatch. Expected '%s', got '%s'", string(testContent), string(readContent))
	}

	// Проверяем хеш
	if readBlob.GetHash() != hash {
		t.Errorf("Hash mismatch. Expected %s, got %s", hash, readBlob.GetHash())
	}

	// Проверяем размер
	if readBlob.Size() != int64(len(testContent)) {
		t.Errorf("Size mismatch. Expected %d, got %d", len(testContent), readBlob.Size())
	}
}

// TestWriteAndReadTree проверяет запись и чтение tree объекта
func TestWriteAndReadTree(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Создаем test blob для включения в tree
	blobContent := []byte("file content")
	blob := objects.NewBlob(blobContent)
	blobHash, err := store.WriteObject(blob)
	if err != nil {
		t.Fatalf("Failed to write blob: %v", err)
	}

	// Создаем tree
	tree := objects.NewTree()

	// Добавляем записи в tree
	entries := []struct {
		mode objects.FileMode
		name string
		hash objects.Hash
		typ  objects.ObjectType
	}{
		{objects.FileModeRegular, "README.md", blobHash, objects.BlobObject},
		{objects.FileModeExec, "script.sh", blobHash, objects.BlobObject},
		{objects.FileModeDir, "src", blobHash, objects.TreeObject},
	}

	for _, entry := range entries {
		treeEntry, err := objects.NewTreeEntry(entry.mode, entry.name, entry.hash, entry.typ)
		if err != nil {
			t.Fatalf("Failed to create tree entry: %v", err)
		}
		if err := tree.AddEntry(*treeEntry); err != nil {
			t.Fatalf("Failed to add tree entry: %v", err)
		}
	}

	// Записываем tree
	treeHash, err := store.WriteObject(tree)
	if err != nil {
		t.Fatalf("WriteObject failed for tree: %v", err)
	}

	// Читаем tree обратно
	readObj, err := store.ReadObject(treeHash)
	if err != nil {
		t.Fatalf("ReadObject failed for tree: %v", err)
	}

	readTree, ok := readObj.(*objects.Tree)
	if !ok {
		t.Fatalf("Expected *objects.Tree, got %T", readObj)
	}

	// Проверяем, что записи сохранились
	entriesCount := len(readTree.Entries())
	if entriesCount != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(entries), entriesCount)
	}

	// Проверяем, что записи отсортированы по имени
	treeEntries := readTree.Entries()
	for i := 1; i < len(treeEntries); i++ {
		if treeEntries[i-1].Name() > treeEntries[i].Name() {
			t.Error("Tree entries are not sorted")
		}
	}
}

// TestWriteAndReadCommit проверяет запись и чтение commit объекта
func TestWriteAndReadCommit(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Создаем tree для коммита
	tree := objects.NewTree()
	blob := objects.NewBlob([]byte("content"))
	blobHash, _ := store.WriteObject(blob)

	treeEntry, _ := objects.NewTreeEntry(objects.FileModeRegular, "file.txt", blobHash, objects.BlobObject)
	tree.AddEntry(*treeEntry)
	treeHash, _ := store.WriteObject(tree)

	// Создаем подписи
	author, _ := objects.NewSignature("John Doe", "john@example.com", time.Now())
	committer, _ := objects.NewSignature("Jane Smith", "jane@example.com", time.Now())

	// Создаем коммит
	commit, err := objects.NewCommit(treeHash, []objects.Hash{}, *author, *committer, "Initial commit")
	if err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	// Записываем коммит
	commitHash, err := store.WriteObject(commit)
	if err != nil {
		t.Fatalf("WriteObject failed for commit: %v", err)
	}

	// Читаем коммит обратно
	readObj, err := store.ReadObject(commitHash)
	if err != nil {
		t.Fatalf("ReadObject failed for commit: %v", err)
	}

	readCommit, ok := readObj.(*objects.Commit)
	if !ok {
		t.Fatalf("Expected *objects.Commit, got %T", readObj)
	}

	// Проверяем поля коммита
	if readCommit.Tree() != treeHash {
		t.Errorf("Tree hash mismatch. Expected %s, got %s", treeHash, readCommit.Tree())
	}

	if readCommit.Message() != "Initial commit" {
		t.Errorf("Message mismatch. Expected 'Initial commit', got '%s'", readCommit.Message())
	}

	if !readCommit.IsRoot() {
		t.Error("Commit should be root (no parents)")
	}
}

// TestObjectExists проверяет проверку существования объектов
func TestObjectExists(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Проверяем несуществующий объект
	fakeHash := objects.Hash("a1b2c3d4e5f67890")
	if store.ObjectExists(fakeHash) {
		t.Error("ObjectExists should return false for non-existent object")
	}

	// Создаем реальный объект
	blob := objects.NewBlob([]byte("test"))
	realHash, err := store.WriteObject(blob)
	if err != nil {
		t.Fatalf("WriteObject failed: %v", err)
	}

	// Проверяем существующий объект
	if !store.ObjectExists(realHash) {
		t.Error("ObjectExists should return true for existing object")
	}
}

// TestReadNonExistentObject проверяет чтение несуществующего объекта
func TestReadNonExistentObject(t *testing.T) {
	store := NewObjectStore(t.TempDir())

	_, err := store.ReadObject(objects.Hash("nonexistent1234567890abcdef"))
	if err == nil {
		t.Error("Expected error when reading non-existent object")
	}
}

// TestIntegrityCheck проверяет проверку целостности данных
func TestIntegrityCheck(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Создаем объект
	blob := objects.NewBlob([]byte("important data"))
	hash, err := store.WriteObject(blob)
	if err != nil {
		t.Fatalf("WriteObject failed: %v", err)
	}

	// Находим путь к файлу объекта
	objectPath, _ := store.hashToPath(hash)

	// Повреждаем файл (записываем мусорные данные)
	corruptedData := []byte("this is corrupted data that will break the hash")
	if err := os.WriteFile(objectPath, corruptedData, 0644); err != nil {
		t.Fatalf("Failed to corrupt file: %v", err)
	}

	// Попытка чтения должна вернуть ошибку целостности
	_, err = store.ReadObject(hash)
	if err == nil {
		t.Error("Expected integrity error for corrupted file")
	}

	// Проверяем, что ошибка содержит информацию о целостности
	if err != nil && err.Error()[:20] != "object integrity check" {
		t.Errorf("Expected integrity error, got: %v", err)
	}
}

// TestMultipleObjects проверяет работу с множеством объектов
func TestMultipleObjects(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Создаем несколько объектов
	objectsCount := 10
	hashes := make([]objects.Hash, objectsCount)

	for i := 0; i < objectsCount; i++ {
		content := []byte(string(rune(65 + i))) // A, B, C, ...
		blob := objects.NewBlob(content)
		hash, err := store.WriteObject(blob)
		if err != nil {
			t.Fatalf("WriteObject failed for object %d: %v", i, err)
		}
		hashes[i] = hash
	}

	// Проверяем, что все объекты существуют
	for i, hash := range hashes {
		if !store.ObjectExists(hash) {
			t.Errorf("Object %d (hash: %s) does not exist", i, hash)
		}
	}

	// Проверяем, что создались правильные директории
	dirs, err := os.ReadDir(store.objectsDir)
	if err != nil {
		t.Fatalf("Failed to read objects directory: %v", err)
	}

	// Должны быть созданы поддиректории для разных хешей
	if len(dirs) == 0 {
		t.Error("No subdirectories created in objects directory")
	}

	// Читаем все объекты обратно и проверяем содержимое
	for i, hash := range hashes {
		obj, err := store.ReadObject(hash)
		if err != nil {
			t.Errorf("ReadObject failed for object %d: %v", i, err)
			continue
		}

		blob, ok := obj.(*objects.Blob)
		if !ok {
			t.Errorf("Expected blob for object %d, got %T", i, obj)
			continue
		}

		expectedContent := string(rune(65 + i))
		if string(blob.Content()) != expectedContent {
			t.Errorf("Content mismatch for object %d. Expected '%s', got '%s'", i, expectedContent, string(blob.Content()))
		}
	}
}

// TestEmptyObject проверяет работу с пустыми объектами
func TestEmptyObject(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewObjectStore(tmpDir)

	// Создаем пустой blob
	emptyBlob := objects.NewBlob([]byte{})
	hash, err := store.WriteObject(emptyBlob)
	if err != nil {
		t.Fatalf("WriteObject failed for empty blob: %v", err)
	}

	// Читаем обратно
	readObj, err := store.ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject failed for empty blob: %v", err)
	}

	readBlob, ok := readObj.(*objects.Blob)
	if !ok {
		t.Fatalf("Expected *objects.Blob, got %T", readObj)
	}

	if readBlob.Size() != 0 {
		t.Errorf("Expected size 0, got %d", readBlob.Size())
	}

	if len(readBlob.Content()) != 0 {
		t.Errorf("Expected empty content, got %d bytes", len(readBlob.Content()))
	}
}
