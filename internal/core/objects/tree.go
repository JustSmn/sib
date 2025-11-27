package objects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// TreeEntry представляет запись в tree объекте (файл или поддиректория)
type TreeEntry struct {
	mode    FileMode   // Приватно: режим файла
	name    string     // Приватно: имя файла/директории
	hash    Hash       // Приватно: хеш объекта
	objType ObjectType // Приватно: тип объекта
}

func NewTreeEntry(mode FileMode, name string, hash Hash, objType ObjectType) (*TreeEntry, error) {
	if err := mode.Validate(); err != nil {
		return nil, fmt.Errorf("invalid file mode in tree entry: %w", err)
	}
	if name == "" {
		return nil, fmt.Errorf("tree entry name cannot be empty")
	}
	if hash.IsEmpty() {
		return nil, fmt.Errorf("tree entry hash cannot be empty")
	}
	if err := objType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid object type in tree entry: %w", err)
	}

	return &TreeEntry{
		mode:    mode,
		name:    name,
		hash:    hash,
		objType: objType,
	}, nil
}

// Mode возвращает режим файла
func (te *TreeEntry) Mode() FileMode { return te.mode }

// Name возвращает имя файла/директории
func (te *TreeEntry) Name() string { return te.name }

// Hash возвращает хеш объекта
func (te *TreeEntry) Hash() Hash { return te.hash }

// Type возвращает тип объекта
func (te *TreeEntry) Type() ObjectType { return te.objType }

// Tree представляет структуру каталога
// Tree содержит список записей (файлов и поддиректорий)
type Tree struct {
	entries []TreeEntry // Приватно: гарантия сортировки и контроля доступа
	hash    Hash        // Приватно: хеш объекта
}

// NewTree создает новый пустой tree
func NewTree() *Tree {
	return &Tree{
		entries: make([]TreeEntry, 0),
	}
}

func (t *Tree) AddEntry(entry TreeEntry) error {
	// TODO: РЕАЛИЗОВАТЬ ВАЛИДАЦИЮ

	// Проверяем дубликаты по имени
	for i, e := range t.entries {
		if e.name == entry.name {

			// Обновляем существующую запись
			t.entries[i] = entry
			t.sortEntries()

			return nil
		}
	}

	// Добавляем новую запись
	t.entries = append(t.entries, entry)
	t.sortEntries()

	return nil
}

// RemoveEntry удаляет запись по имени
func (t *Tree) RemoveEntry(name string) bool {

	for i, entry := range t.entries {
		if entry.name == name {

			// Удаляем запись, сохраняя порядок
			t.entries = append(t.entries[:i], t.entries[i+1:]...)

			return true
		}
	}
	return false
}

// GetEntry возвращает запись по имени
func (t *Tree) GetEntry(name string) (*TreeEntry, bool) {

	for _, entry := range t.entries {
		if entry.name == name {
			return &entry, true
		}
	}
	return nil, false
}

// Entries возвращает копию списка записей (для иммутабельности)
func (t *Tree) Entries() []TreeEntry {
	entriesCopy := make([]TreeEntry, len(t.entries))
	copy(entriesCopy, t.entries)

	return entriesCopy
}

// sortEntries сортирует записи по имени в лексикографическом порядке
// Это критично для детерминированных хешей
func (t *Tree) sortEntries() {
	sort.Slice(t.entries, func(i, j int) bool {
		return t.entries[i].name < t.entries[j].name
	})
}

// Hash возвращает хеш объекта
func (t *Tree) Hash() Hash { return t.hash }

// SetHash устанавливает хеш объекта
func (t *Tree) SetHash(h Hash) { t.hash = h }

// Type возвращает тип объекта
func (t *Tree) Type() ObjectType { return TreeObject }

// Serialize преобразует tree в байтовое представление
// Формат: канонический JSON с отсортированными полями
func (t *Tree) Serialize() ([]byte, error) {
	if len(t.entries) == 0 {
		return nil, fmt.Errorf("tree cannot be empty")
	}

	// Создаем структуру для сериализации с гарантированным порядком полей
	type serializableTree struct {
		Type    ObjectType  `json:"type"`
		Entries []TreeEntry `json:"entries"`
	}

	st := serializableTree{
		Type:    TreeObject,
		Entries: t.entries,
	}

	// Используем канонический JSON (отсортированные ключи, без лишних пробелов)
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "") // Без отступов для детерминизма

	if err := encoder.Encode(st); err != nil {
		return nil, fmt.Errorf("failed to serialize tree: %w", err)
	}

	// Убираем лишний перенос строки, добавленный json.Encoder
	data := bytes.TrimSpace(buf.Bytes())

	// Добавляем Git-заголовок
	header := fmt.Sprintf("%s %d", t.Type(), len(data))

	result := append([]byte(header), 0)
	result = append(result, data...)

	return result, nil
}

// DeserializeTree создает Tree из байтового представления
// Это пригодится при чтении объектов из хранилища
// DeserializeTree создает Tree из байтового представления
func DeserializeTree(data []byte) (*Tree, error) {
	// Находим разделитель
	for i, b := range data {
		if b == 0 {
			// Парсим JSON часть
			var st struct {
				Type    ObjectType  `json:"type"`
				Entries []TreeEntry `json:"entries"`
			}

			if err := json.Unmarshal(data[i+1:], &st); err != nil {
				return nil, fmt.Errorf("failed to deserialize tree: %w", err)
			}

			if st.Type != TreeObject {
				return nil, fmt.Errorf("invalid object type: expected tree, got %s", st.Type)
			}

			// Создаем tree и добавляем записи
			tree := NewTree()
			for _, entry := range st.Entries {
				if err := tree.AddEntry(entry); err != nil {
					return nil, fmt.Errorf("invalid tree entry during deserialization: %w", err)
				}
			}

			return tree, nil
		}
	}

	return nil, fmt.Errorf("malformed tree data: no null byte separator")
}
