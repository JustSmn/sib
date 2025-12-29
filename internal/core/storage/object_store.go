package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"sib/internal/core/objects"
	"sib/internal/utils"
)

// ObjectStore представляет CAS-хранилище объектов
type ObjectStore struct {
	objectsDir string // Путь к директории objects (например, .sib/objects)
}

// NewObjectStore создает новое хранилище объектов
func NewObjectStore(repoPath string) (*ObjectStore, error) {
	objectsDir := filepath.Join(repoPath, ".sib", "objects")

	// ПРОВЕРЯЕМ, что директория существует
	if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a sib repository: .sib/objects not found")
	}

	return &ObjectStore{
		objectsDir: objectsDir,
	}, nil
}

// calculateHash вычисляет SHA-256 хеш от данных
func (store *ObjectStore) calculateHash(data []byte) objects.Hash {
	hash := sha256.Sum256(data)

	// Преобразуем [32]byte в строку в hex-формате
	return objects.Hash(fmt.Sprintf("%x", hash))
}

// hashToPath преобразует хеш в путь к файлу в структуре objects/ab/cdef...
func (store *ObjectStore) hashToPath(hash objects.Hash) (string, error) {
	hashStr := hash.String()
	if len(hashStr) < 2 {
		return "", fmt.Errorf("hash too short: %s", hash)
	}

	// Берем первые 2 символа для директории, остальные для имени файла
	dirName := hashStr[:2]
	fileName := hashStr[2:]

	return filepath.Join(store.objectsDir, dirName, fileName), nil
}

// WriteObject сохраняет объект в CAS-хранилище
// Возвращает хеш объекта или ошибку, если что-то пошло не так
func (store *ObjectStore) WriteObject(obj objects.Serializable) (objects.Hash, error) {
	// Сериализуем объект в байты
	data, err := obj.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize object: %w", err)
	}

	// Вычисляем SHA-256 хеш от сериализованных данных
	hash := store.calculateHash(data)

	// Преобразуем хеш в путь к файлу (структура ab/cdef...)
	objectPath, err := store.hashToPath(hash)
	if err != nil {
		return "", fmt.Errorf("failed to create object path: %w", err)
	}

	// Создаем директорию, если её нет (только для первых двух символов хеша)
	dir := filepath.Dir(objectPath)
	if err := utils.CreateDirIfNotExists(dir); err != nil {
		return "", fmt.Errorf("failed to create object directory: %w", err)
	}

	// Сжимаем данные с помощью Zstd
	compressedData, err := utils.CompressZstd(data)
	if err != nil {
		return "", fmt.Errorf("failed to compress object: %w", err)
	}

	// Атомарно записываем файл (чтобы избежать частичной записи)
	if err := utils.WriteFileAtomic(objectPath, compressedData); err != nil {
		return "", fmt.Errorf("failed to write object file: %w", err)
	}

	// Устанавливаем хеш в объект (если он поддерживает Hashable)
	if hashable, ok := obj.(objects.Hashable); ok {
		hashable.SetHash(hash)
	}

	return hash, nil
}

// ReadObject читает объект из CAS-хранилища по хешу
// Возвращает десериализованный объект или ошибку
func (store *ObjectStore) ReadObject(hash objects.Hash) (objects.Serializable, error) {
	// Проверяем, что хеш не пустой
	if hash.IsEmpty() {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	// Преобразуем хеш в путь к файлу
	objectPath, err := store.hashToPath(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to create object path: %w", err)
	}

	// Читаем сжатые данные из файла
	compressedData, err := utils.ReadFile(objectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object file: %w", err)
	}

	// Декомпрессируем данные
	data, err := utils.DecompressZstd(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress object: %w", err)
	}

	// Проверяем целостность: вычисляем хеш заново и сравниваем
	calculatedHash := store.calculateHash(data)
	if calculatedHash != hash {
		return nil, fmt.Errorf("object integrity check failed: expected %s, got %s", hash, calculatedHash)
	}

	// Определяем тип объекта и десериализуем
	objType, err := store.detectObjectType(data)
	if err != nil {
		return nil, fmt.Errorf("failed to detect object type: %w", err)
	}

	// Десериализуем объект в зависимости от типа
	obj, err := store.deserializeByType(objType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object: %w", err)
	}

	// Устанавливаем хеш в объект
	if hashable, ok := obj.(objects.Hashable); ok {
		hashable.SetHash(hash)
	}

	return obj, nil
}

// detectObjectType определяет тип объекта по его сериализованным данным
func (store *ObjectStore) detectObjectType(data []byte) (objects.ObjectType, error) {
	// Ищем нулевой байт-разделитель между заголовком и содержимым
	for i, b := range data {
		if b == 0 {
			header := string(data[:i])
			var objType objects.ObjectType
			var size int

			// Парсим заголовок в формате "type size"
			if _, err := fmt.Sscanf(header, "%s %d", &objType, &size); err != nil {
				return "", fmt.Errorf("failed to parse object header: %w", err)
			}

			if err := objType.Validate(); err != nil {
				return "", fmt.Errorf("invalid object type in header: %w", err)
			}

			return objType, nil
		}
	}

	return "", fmt.Errorf("object data malformed: no null byte separator found")
}

// deserializeByType десериализует объект в зависимости от его типа
func (store *ObjectStore) deserializeByType(objType objects.ObjectType, data []byte) (objects.Serializable, error) {
	switch objType {
	case objects.BlobObject:
		// Для Blob просто создаем из данных после заголовка
		for i, b := range data {
			if b == 0 {
				return objects.NewBlob(data[i+1:]), nil
			}
		}
		return nil, fmt.Errorf("malformed blob data")

	case objects.TreeObject:
		return objects.DeserializeTree(data)

	case objects.CommitObject:
		return objects.DeserializeCommit(data)

	case objects.TagObject:
		// TODO: реализовать DeserializeTag когда будешь делать теги
		return nil, fmt.Errorf("tag deserialization not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported object type: %s", objType)
	}
}

// ObjectExists проверяет, существует ли объект с указанным хешом
func (store *ObjectStore) ObjectExists(hash objects.Hash) bool {
	objectPath, err := store.hashToPath(hash)
	if err != nil {
		return false
	}
	return utils.FileExists(objectPath)
}
