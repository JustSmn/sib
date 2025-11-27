package objects

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Tag представляет аннотированный тег
type Tag struct {
	object  Hash       // Приватно: хеш объекта, на который ссылается тег
	objType ObjectType // Приватно: тип объекта
	tagName string     // Приватно: имя тега
	tagger  Signature  // Приватно: автор тега
	message string     // Приватно: сообщение тега
	hash    Hash       // Приватно: хеш самого тега
}

// NewTag создает новый тег с валидацией
func NewTag(object Hash, objType ObjectType, tagName string, tagger Signature, message string) (*Tag, error) {
	if object.IsEmpty() {
		return nil, fmt.Errorf("tag object cannot be empty")
	}
	if err := objType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid object type: %w", err)
	}
	if tagName == "" {
		return nil, fmt.Errorf("tag name cannot be empty")
	}
	if err := tagger.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tagger signature: %w", err)
	}

	return &Tag{
		object:  object,
		objType: objType,
		tagName: tagName,
		tagger:  tagger,
		message: message,
	}, nil
}

// Object возвращает хеш объекта, на который ссылается тег
func (t *Tag) Object() Hash { return t.object }

// ObjectType возвращает тип объекта, на который ссылается тег
func (t *Tag) ObjectType() ObjectType { return t.objType }

// TagName возвращает имя тега (например, "v1.0.0")
func (t *Tag) TagName() string { return t.tagName }

// Tagger возвращает подпись автора тега
func (t *Tag) Tagger() Signature { return t.tagger }

// Message возвращает сообщение тега (описание)
func (t *Tag) Message() string { return t.message }

// Hash возвращает хеш объекта
func (t *Tag) Hash() Hash { return t.hash }

// SetHash устанавливает хеш объекта
func (t *Tag) SetHash(h Hash) { t.hash = h }

// Type возвращает тип объекта
func (t *Tag) Type() ObjectType { return TagObject }

// Serialize преобразует tag в байтовое представление
func (t *Tag) Serialize() ([]byte, error) {
	type serializableTag struct {
		Type    ObjectType `json:"type"`
		Object  Hash       `json:"object"`
		ObjType ObjectType `json:"objType"`
		Tag     string     `json:"tag"`
		Tagger  Signature  `json:"tagger"`
		Message string     `json:"message"`
	}

	st := serializableTag{
		Type:    TagObject,
		Object:  t.object,
		ObjType: t.objType,
		Tag:     t.tagName,
		Tagger:  t.tagger,
		Message: t.message,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	if err := encoder.Encode(st); err != nil {
		return nil, fmt.Errorf("failed to serialize tag: %w", err)
	}

	data := bytes.TrimSpace(buf.Bytes())
	header := fmt.Sprintf("%s %d", t.Type(), len(data))
	result := append([]byte(header), 0)
	result = append(result, data...)

	return result, nil
}
