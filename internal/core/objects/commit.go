package objects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Commit представляет коммит в истории проекта
type Commit struct {
	tree      Hash      // Приватно: хеш корневого tree
	parents   []Hash    // Приватно: хеши родительских коммитов
	author    Signature // Приватно: автор изменений
	committer Signature // Приватно: тот, кто создал коммит
	message   string    // Приватно: сообщение коммита
	hash      Hash      // Приватно: хеш самого коммита
}

// NewCommit создает новый коммит с валидацией
func NewCommit(tree Hash, parents []Hash, author, committer Signature, message string) (*Commit, error) {
	if tree.IsEmpty() {
		return nil, fmt.Errorf("commit tree cannot be empty")
	}
	if err := author.Validate(); err != nil {
		return nil, fmt.Errorf("invalid author: %w", err)
	}
	if err := committer.Validate(); err != nil {
		return nil, fmt.Errorf("invalid committer: %w", err)
	}
	if strings.TrimSpace(message) == "" { //Так мы проверяем, что в сообщении коммита есть реальный текст, а не только пробелы.
		return nil, fmt.Errorf("commit message cannot be empty")
	}
	//strings.TrimSpace(message) - это функция, которая обрезает пробелы и переносы строк в начале и конце строки.
	//Например:
	//" hello " → "hello"
	//"\n\ncommit message\n" → "commit message"
	//" " → "" (пустая строка)

	return &Commit{
		tree:      tree,
		parents:   parents,
		author:    author,
		committer: committer,
		message:   strings.TrimSpace(message),
	}, nil
}

// Tree возвращает хеш корневого tree
func (c *Commit) Tree() Hash { return c.tree }

// Parents возвращает копию списка родительских коммитов
func (c *Commit) Parents() []Hash {
	parentsCopy := make([]Hash, len(c.parents))
	copy(parentsCopy, c.parents)

	return parentsCopy
}

// Author возвращает подпись автора
func (c *Commit) Author() Signature { return c.author }

// Committer возвращает подпись коммитера
func (c *Commit) Committer() Signature { return c.committer }

// Message возвращает сообщение коммита
func (c *Commit) Message() string { return c.message }

// IsMerge проверяет, является ли коммит слиянием
func (c *Commit) IsMerge() bool { return len(c.parents) >= 2 }

// IsRoot проверяет, является ли коммит корневым (без родителей)
func (c *Commit) IsRoot() bool { return len(c.parents) == 0 }

// Hash возвращает хеш объекта
func (c *Commit) Hash() Hash { return c.hash }

// SetHash устанавливает хеш объекта
func (c *Commit) SetHash(h Hash) { c.hash = h }

// Type возвращает тип объекта
func (c *Commit) Type() ObjectType { return CommitObject }

/*
// Serialize преобразует commit в байтовое представление
// Формат: канонический JSON
func (c *Commit) Serialize() ([]byte, error) {
	// Структура для сериализации
	type serializableCommit struct {
		Type      ObjectType `json:"type"`
		Tree      Hash       `json:"tree"`
		Parents   []Hash     `json:"parents,omitempty"` //omitempty = если поле пустое (empty), не включай его в JSON".
		Author    Signature  `json:"author"`
		Committer Signature  `json:"committer"`
		Message   string     `json:"message"`
		Timestamp int64      `json:"timestamp"` // Unix timestamp для детерминизма
	}

	// Используем Unix timestamp для детерминированной сериализации времени
	timestamp := c.author.Time().Unix()

	sc := serializableCommit{
		Type:      CommitObject,
		Tree:      c.tree,
		Parents:   c.parents,
		Author:    c.author,
		Committer: c.committer,
		Message:   c.message,
		Timestamp: timestamp,
	}

	// Канонический JSON
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	if err := encoder.Encode(sc); err != nil {
		return nil, fmt.Errorf("failed to serialize commit: %w", err)
	}

	data := bytes.TrimSpace(buf.Bytes())

	// Git-заголовок
	header := fmt.Sprintf("%s %d", c.Type(), len(data))

	result := append([]byte(header), 0)
	result = append(result, data...)

	return result, nil
}

// DeserializeCommit создает Commit из байтового представления
func DeserializeCommit(data []byte) (*Commit, error) {
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid commit data format")
	}

	var sc struct {
		Type      ObjectType `json:"type"`
		Tree      Hash       `json:"tree"`
		Parents   []Hash     `json:"parents"`
		Author    Signature  `json:"author"`
		Committer Signature  `json:"committer"`
		Message   string     `json:"message"`
		Timestamp int64      `json:"timestamp"`
	}

	if err := json.Unmarshal(parts[1], &sc); err != nil {
		return nil, fmt.Errorf("failed to deserialize commit: %w", err)
	}

	if sc.Type != CommitObject {
		return nil, fmt.Errorf("invalid object type: expected commit, got %s", sc.Type)
	}

	// Восстанавливаем время из timestamp
	authorTime := time.Unix(sc.Timestamp, 0)
	author, err := NewSignature(sc.Author.Name(), sc.Author.Email(), authorTime)
	if err != nil {
		return nil, fmt.Errorf("invalid author signature: %w", err)
	}

	committer, err := NewSignature(sc.Committer.Name(), sc.Committer.Email(), authorTime)
	if err != nil {
		return nil, fmt.Errorf("invalid committer signature: %w", err)
	}

	commit, err := NewCommit(sc.Tree, sc.Parents, *author, *committer, sc.Message)
	if err != nil {
		return nil, fmt.Errorf("deserialized commit validation failed: %w", err)
	}

	return commit, nil
}
*/

// ==================== СЕРИАЛИЗАЦИЯ ДЛЯ COMMIT ====================

// signatureJSON - приватная структура для JSON сериализации
type signatureJSON struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	When  time.Time `json:"when"`
}

// commitJSON - приватная структура для JSON сериализации
type commitJSON struct {
	Type      ObjectType    `json:"type"`
	Tree      Hash          `json:"tree"`
	Parents   []Hash        `json:"parents,omitempty"`
	Author    signatureJSON `json:"author"`
	Committer signatureJSON `json:"committer"`
	Message   string        `json:"message"`
	Timestamp int64         `json:"timestamp"`
}

// toJSONSignature конвертирует Signature в signatureJSON
func (s *Signature) toJSONSignature() signatureJSON {
	return signatureJSON{
		Name:  s.name,
		Email: s.email,
		When:  s.when,
	}
}

// fromJSONSignature создает Signature из signatureJSON
func fromJSONSignature(sj signatureJSON) (*Signature, error) {
	return NewSignature(sj.Name, sj.Email, sj.When)
}

// Serialize преобразует commit в байтовое представление
func (c *Commit) Serialize() ([]byte, error) {
	// Конвертируем в JSON формат
	cj := commitJSON{
		Type:      CommitObject,
		Tree:      c.tree,
		Parents:   c.parents,
		Author:    c.author.toJSONSignature(),
		Committer: c.committer.toJSONSignature(),
		Message:   c.message,
		Timestamp: c.author.Time().Unix(),
	}

	// Канонический JSON
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	if err := encoder.Encode(cj); err != nil {
		return nil, fmt.Errorf("failed to serialize commit: %w", err)
	}

	data := bytes.TrimSpace(buf.Bytes())
	header := fmt.Sprintf("%s %d", c.Type(), len(data))
	result := append([]byte(header), 0)
	result = append(result, data...)

	return result, nil
}

// DeserializeCommit создает Commit из байтового представления
func DeserializeCommit(data []byte) (*Commit, error) {
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid commit data format")
	}

	var cj commitJSON
	if err := json.Unmarshal(parts[1], &cj); err != nil {
		return nil, fmt.Errorf("failed to deserialize commit: %w", err)
	}

	if cj.Type != CommitObject {
		return nil, fmt.Errorf("invalid object type: expected commit, got %s", cj.Type)
	}

	// Восстанавливаем Signature из JSON
	author, err := fromJSONSignature(cj.Author)
	if err != nil {
		return nil, fmt.Errorf("invalid author signature: %w", err)
	}

	committer, err := fromJSONSignature(cj.Committer)
	if err != nil {
		return nil, fmt.Errorf("invalid committer signature: %w", err)
	}

	// Создаем commit
	commit, err := NewCommit(cj.Tree, cj.Parents, *author, *committer, cj.Message)
	if err != nil {
		return nil, fmt.Errorf("deserialized commit validation failed: %w", err)
	}

	return commit, nil
}
