package objects

import "fmt"

// Blob представляет содержимое файла
// В Git blob хранит только данные файла, без имени и пути
type Blob struct {
	content []byte // Приватно: Содержимое файла в бинарном виде
	size    int64  // Приватно: Размер содержимого в байтах (вычисляется автоматически)
	hash    Hash   // Приватно: SHA-256 хеш объекта (устанавливается после сохранения)
}

// NewBlob создает новый blob из переданного содержимого
func NewBlob(content []byte) *Blob {
	return &Blob{
		content: content,
		size:    int64(len(content)),
		// Hash не устанавливаем здесь, потому что:
		// - Хеш вычисляется после сериализации объекта
		// - Хеш зависит от полных данных (заголовок + содержимое)
		// - Хеш устанавливается в CAS-хранилище после успешного сохранения
		// - Если установить хеш заранее - он будет неверным
	}
}

// Content возвращает копию содержимого blob
// Возвращаем копию для защиты от изменений исходных данных
func (b *Blob) Content() []byte {
	contentCopy := make([]byte, len(b.content))
	copy(contentCopy, b.content)

	return contentCopy
}

// Size возвращает размер содержимого blob в байтах
func (b *Blob) Size() int64 {
	return b.size
}

// Hash возвращает хеш объекта
func (b *Blob) GetHash() Hash {
	return b.hash
}

// SetHash устанавливает хеш объекта
// Вызывается CAS-хранилищем после успешного сохранения
func (b *Blob) SetHash(h Hash) {
	b.hash = h
}

// Возвращает тип объекта
func (b *Blob) Type() ObjectType {
	return BlobObject
}

// Serialize преобразует blob в байтовое представление для хеширования
// Формат: "blob <размер>\0<содержимое>"
// "\0" - это не строка, а нулевой байт разделитель
// Возвращает ошибку если: размер отрицательный, содержимое nil
func (b *Blob) Serialize() ([]byte, error) {
	if b.size < 0 {
		return nil, fmt.Errorf("invalid blob size: %d", b.size)
	}
	if b.content == nil {
		return nil, fmt.Errorf("blob content is nil")
	}

	// Проверяем соответствие размера и фактического содержимого
	actualSize := int64(len(b.content))
	if b.size != actualSize {
		return nil, fmt.Errorf("size mismatch: declared %d, actual %d", b.size, actualSize)
	}

	// Создаем заголовок в формате Git
	header := fmt.Sprintf("%s %d", b.Type(), b.size)

	data := append([]byte(header), 0) // Это байт со значением 0, а не строка "\0"
	data = append(data, b.content...)

	return data, nil
}
