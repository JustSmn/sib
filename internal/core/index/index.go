package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type IndexEntry struct {
	// Обязательные поля (должны быть в JSON):
	Hash  string    `json:"hash"`  // SHA-256 хеш содержимого файла
	Size  int64     `json:"size"`  // Размер файла в байтах
	Mode  string    `json:"mode"`  // Права доступа: "100644", "100755", "040000"
	Mtime time.Time `json:"mtime"` // Время последнего изменения файла
	Path  string    `json:"path"`  // Относительный путь (от корня репозитория)

	// Служебные поля (не обязательны в JSON):
	ctime     time.Time // Время создания записи в индексе (для отладки)
	validated bool      // Проверен ли файл на целостность
	stage     int       // Стадия: 0 = нормальная, 1-3 = конфликт слияния
	/*
	   Почему такие поля:
	     Hash — найти содержимое в CAS-хранилище
	     Size + Mtime — быстро понять, изменился ли файл (без чтения всего)
	     Mode — важно для исполняемых файлов
	     Path — где восстановить файл при checkout*/
}

type Index struct {
	// Приватные поля:
	path    string // Путь к файлу .sib/index
	version int    // Версия формата (начинаем с 1)

	// Публичные (для JSON):
	Version int                   `json:"version"` // Версия формата
	Entries map[string]IndexEntry `json:"entries"` // Ключ: путь к файлу
}

// NewIndex создает или загружает индекс из файла
func NewIndex(repoPath string) (*Index, error) {
	// Формируем путь к файлу индекса
	indexPath := filepath.Join(repoPath, ".sib", "index")

	// Создаем директорию .sib если её нет
	sibDir := filepath.Join(repoPath, ".sib")
	if err := os.MkdirAll(sibDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .sib directory: %w", err)
	}

	// Создаем сам индекс (файл) если его нет
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Создаем пустой файл индекса
		emptyIndex := &Index{
			path:    indexPath,
			version: 1,
			Version: 1,
			Entries: make(map[string]IndexEntry),
		}
		if err := emptyIndex.Save(); err != nil {
			return nil, fmt.Errorf("failed to create index file: %w", err)
		}
	}

	// Дальше загружаем существующий индекс...
	idx := &Index{
		path:    indexPath,
		version: 1,
		Version: 1,
		Entries: make(map[string]IndexEntry),
	}

	if err := idx.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return idx, nil
}

// load загружает индекс из файла (приватный метод)
func (idx *Index) load() error {
	// Читаем файл
	data, err := os.ReadFile(idx.path)
	if err != nil {
		return err
	}

	// Проверяем, не пустой ли файл
	if len(data) == 0 {
		idx.Entries = make(map[string]IndexEntry)
		return nil
	}

	// Парсим JSON
	var loadedIndex struct {
		Version int                   `json:"version"`
		Entries map[string]IndexEntry `json:"entries"`
	}

	if err := json.Unmarshal(data, &loadedIndex); err != nil {
		// Если JSON поврежден, создаем новый пустой индекс
		idx.Entries = make(map[string]IndexEntry)
		return nil
	}

	// Копируем загруженные данные
	idx.Version = loadedIndex.Version
	idx.Entries = loadedIndex.Entries

	return nil
}

// Save сохраняет индекс в файл
func (idx *Index) Save() error {
	// Сортируем ключи для детерминированного JSON
	keys := make([]string, 0, len(idx.Entries))
	for key := range idx.Entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Создаем упорядоченную мапу для сериализации
	orderedEntries := make(map[string]IndexEntry)
	for _, key := range keys {
		orderedEntries[key] = idx.Entries[key]
	}
	idx.Entries = orderedEntries

	// Сериализуем в JSON с отступами
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	// Атомарная запись через временный файл
	tmpPath := idx.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary index file: %w", err)
	}

	// Атомарное переименование
	if err := os.Rename(tmpPath, idx.path); err != nil {
		// Пытаемся удалить временный файл в случае ошибки
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename index file: %w", err)
	}

	return nil
}

// Add добавляет или обновляет файл в индексе
func (idx *Index) Add(path string, hash string, size int64, mode string, mtime time.Time) error {
	// Валидация входных данных
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}
	if size < 0 {
		return fmt.Errorf("size cannot be negative")
	}

	// Проверяем валидность режима файла
	if !isValidMode(mode) {
		return fmt.Errorf("invalid file mode: %s", mode)
	}

	// Нормализуем путь
	normalizedPath := normalizePath(path)

	// Создаем запись
	entry := IndexEntry{
		Hash:      hash,
		Size:      size,
		Mode:      mode,
		Mtime:     mtime,
		Path:      normalizedPath,
		ctime:     time.Now(),
		validated: true,
		stage:     0,
	}

	// Добавляем в мапу
	idx.Entries[normalizedPath] = entry

	return nil
}

// Remove удаляет файл из индекса
func (idx *Index) Remove(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Нормализуем путь
	normalizedPath := normalizePath(path)

	// Проверяем, существует ли запись
	if _, exists := idx.Entries[normalizedPath]; !exists {
		return fmt.Errorf("file not found in index: %s", path)
	}

	// Удаляем запись
	delete(idx.Entries, normalizedPath)

	return nil
}

// Clear очищает индекс (удаляет все записи)
func (idx *Index) Clear() error {
	idx.Entries = make(map[string]IndexEntry)
	return nil
}

// Get возвращает запись по пути
func (idx *Index) Get(path string) (IndexEntry, error) {
	if path == "" {
		return IndexEntry{}, fmt.Errorf("path cannot be empty")
	}

	normalizedPath := normalizePath(path)

	entry, exists := idx.Entries[normalizedPath]
	if !exists {
		return IndexEntry{}, fmt.Errorf("file not found in index: %s", path)
	}

	return entry, nil
}

// HasChanges проверяет, есть ли изменения в индексе
func (idx *Index) HasChanges() bool {
	return len(idx.Entries) > 0
}

// GetAllEntries возвращает все записи в отсортированном порядке
func (idx *Index) GetAllEntries() []IndexEntry {
	// Собираем все записи
	entries := make([]IndexEntry, 0, len(idx.Entries))
	for _, entry := range idx.Entries {
		entries = append(entries, entry)
	}

	// Сортируем по пути для консистентности
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	return entries
}

// Validate проверяет целостность файлов в индексе
func (idx *Index) Validate(repoPath string) ([]string, error) {
	var invalidFiles []string

	for path, entry := range idx.Entries {
		fullPath := filepath.Join(repoPath, path)

		// Проверяем существование файла
		info, err := os.Stat(fullPath)
		if err != nil {
			invalidFiles = append(invalidFiles, path)
			continue
		}

		// Проверяем размер
		if info.Size() != entry.Size {
			invalidFiles = append(invalidFiles, path)
			continue
		}

		// Проверяем время модификации
		// Допускаем небольшую погрешность в 1 секунду из-за округлений
		diff := info.ModTime().Sub(entry.Mtime)
		if diff < -time.Second || diff > time.Second {
			invalidFiles = append(invalidFiles, path)
		}
	}

	if len(invalidFiles) > 0 {
		return invalidFiles, fmt.Errorf("%d files in index are invalid or have been modified", len(invalidFiles))
	}

	return nil, nil
}

// Diff сравнивает индекс с рабочим каталогом
// Возвращает: новые файлы, измененные файлы, удаленные файлы
func (idx *Index) Diff(repoPath string) (added []string, modified []string, deleted []string, err error) {
	// Получаем все файлы в рабочем каталоге
	workingFiles := make(map[string]os.FileInfo)

	// Рекурсивно обходим директорию
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Пропускаем директорию .sib
		if info.IsDir() && info.Name() == ".sib" {
			return filepath.SkipDir
		}

		// Получаем относительный путь
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil // Пропускаем ошибки
		}

		// Пропускаем корневую директорию
		if relPath == "." {
			return nil
		}

		// Добавляем только файлы (не директории)
		if !info.IsDir() {
			normalizedPath := normalizePath(relPath)
			workingFiles[normalizedPath] = info
		}

		return nil
	})

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to scan working directory: %w", err)
	}

	// Находим добавленные файлы (есть в рабочем каталоге, нет в индексе)
	for path := range workingFiles {
		if _, exists := idx.Entries[path]; !exists {
			added = append(added, path)
		}
	}

	// Находим удаленные файлы (есть в индексе, нет в рабочем каталоге)
	for path := range idx.Entries {
		if _, exists := workingFiles[path]; !exists {
			deleted = append(deleted, path)
		}
	}

	// Находим измененные файлы (есть в обоих, но разные метаданные)
	for path, info := range workingFiles {
		if entry, exists := idx.Entries[path]; exists {
			// Проверяем размер
			if info.Size() != entry.Size {
				modified = append(modified, path)
				continue
			}

			// Проверяем время модификации (с учетом погрешности)
			diff := info.ModTime().Sub(entry.Mtime)
			if diff < -time.Second || diff > time.Second {
				modified = append(modified, path)
			}
		}
	}

	// Сортируем результаты для детерминированного вывода
	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(deleted)

	return added, modified, deleted, nil
}

// Path возвращает путь к файлу индекса (для отладки)
func (idx *Index) Path() string {
	return idx.path
}

// Count возвращает количество файлов в индексе
func (idx *Index) Count() int {
	return len(idx.Entries)
}

// GetInvalidFiles возвращает список невалидных файлов без ошибки
func (idx *Index) GetInvalidFiles(repoPath string) []string {
	invalidFiles, _ := idx.Validate(repoPath)
	return invalidFiles
}

// UpdateEntry обновляет существующую запись
func (idx *Index) UpdateEntry(path string, updates map[string]interface{}) error {
	normalizedPath := normalizePath(path)

	entry, exists := idx.Entries[normalizedPath]
	if !exists {
		return fmt.Errorf("file not found in index: %s", path)
	}

	// Применяем обновления
	for key, value := range updates {
		switch key {
		case "validated":
			if validated, ok := value.(bool); ok {
				entry.validated = validated
			}
		case "stage":
			if stage, ok := value.(int); ok {
				entry.stage = stage
			}
		case "ctime":
			if ctime, ok := value.(time.Time); ok {
				entry.ctime = ctime
			}
		}
	}

	idx.Entries[normalizedPath] = entry
	return nil
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// normalizePath нормализует путь для использования в индексе
func normalizePath(path string) string {
	// Очищаем путь
	cleanPath := filepath.Clean(path)

	// Используем forward slash для кросс-платформенности
	return filepath.ToSlash(cleanPath)
}

// isValidMode проверяет валидность режима файла
func isValidMode(mode string) bool {
	// Поддерживаемые режимы файлов в Sib
	validModes := map[string]bool{
		"100644": true, // Обычный файл
		"100755": true, // Исполняемый файл
		"040000": true, // Директория
	}

	return validModes[mode]
}

// DetectFileMode определяет режим файла на основе информации о файле
func DetectFileMode(info os.FileInfo) string {
	if info.IsDir() {
		return "040000"
	}

	if IsExecutable(info) {
		return "100755"
	}

	return "100644"
}

// IsExecutable проверяет, является ли файл исполняемым
func IsExecutable(info os.FileInfo) bool {
	// На Unix-подобных системах проверяем биты разрешений
	mode := info.Mode()

	// Проверяем бит исполняемости для владельца
	if mode&0100 != 0 {
		return true
	}

	// Для Windows проверяем расширение
	ext := filepath.Ext(info.Name())
	executableExts := []string{".exe", ".bat", ".cmd", ".com", ".sh", ".py", ".pl", ".rb"}

	for _, execExt := range executableExts {
		if ext == execExt {
			return true
		}
	}

	return false
}
