package objects

import (
	"fmt"
	"time"
)

// Типы объектов гита
type ObjectType string

const (
	BlobObject   ObjectType = "blob"
	TreeObject   ObjectType = "tree"
	CommitObject ObjectType = "commit"
	TagObject    ObjectType = "tag"
)

// Validate проверяет валидность типа объекта
func (ot ObjectType) Validate() error {
	switch ot {
	case BlobObject, TreeObject, CommitObject, TagObject:
		return nil
	default:
		return fmt.Errorf("invalid object type: %s", ot)
	}
}

// Hash представляет хеш SHA-256 объекта
// Отдельный тип для типобезопасности и предотвращения ошибок
type Hash string

// String преобразует Hash в string для удобства печати и использования
// Реализует интерфейс Stringer из стандартной библиотеки Go
func (h Hash) String() string {
	if h == "" {
		return ""
	}

	return string(h) // Просто преобразует Hash обратно в string
}

// IsEmpty проверяет, пустой ли хеш
func (h Hash) IsEmpty() bool {
	return h == ""
}

// Signature - подпись автора/коммитера
type Signature struct {
	name  string
	email string
	when  time.Time
}

// NewSignature создает новую валидированную подпись
func NewSignature(name, email string, when time.Time) (*Signature, error) {
	if name == "" {
		return nil, fmt.Errorf("signature name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("signature email cannot be empty")
	}
	if when.IsZero() {
		return nil, fmt.Errorf("signature time cannot be zero")
	}

	return &Signature{
		name:  name,
		email: email,
		when:  when,
	}, nil
}

// Name возвращает имя автора
func (s *Signature) Name() string {
	return s.name
}

// Email возвращает email автора
func (s *Signature) Email() string {
	return s.email
}

// Time возвращает время создания
func (s *Signature) Time() time.Time {
	return s.when
}

// Validate проверяет валидность подписи
func (s *Signature) Validate() error {
	if s.name == "" {
		return fmt.Errorf("signature name cannot be empty")
	}
	if s.email == "" {
		return fmt.Errorf("signature email cannot be empty")
	}
	if s.when.IsZero() {
		return fmt.Errorf("signature time cannot be zero")
	}
	return nil
}

// FileMode представляет права доступа и тип файла в формате Git
// Эти значения соответствуют стандартному формату Git для tree объектов
type FileMode string

const (
	FileModeRegular FileMode = "100644" // Обычный файл: rw-r--r--
	FileModeExec    FileMode = "100755" // Исполняемый: rwxr-xr-x
	FileModeDir     FileMode = "40000"  // Директория: drwxr-xr-x
	FileModeSymlink FileMode = "120000" // Символическая ссылка
)

// FileModeRegular = "100644"
// Разбиваем:
// - "100"    = это обычный файл (regular file)
// - "644"    = права доступа: владелец(rw-), группа(r--), другие(r--)
// В Linux: rw-r--r--

// FileModeExec = "100755"
// - "100"    = обычный файл
// - "755"    = права: владелец(rwx), группа(r-x), другие(r-x)
// В Linux: rwxr-xr-x (исполняемый файл)

// FileModeDir = "40000"
// - "40000"  = это директория (в Git так обозначаются папки)
// В Linux: drwxr-xr-x

// Git хранит режимы файлов как строки в tree объектах:
// Пример tree объекта:
// "100644 README.md\0<хеш>40000 src\0<хеш>"
// При создании tree для файла README.md:
// tree.AddEntry("README.md", FileModeRegular, BlobObject, hash)

// Validate проверяет валидность режима файла
func (fm FileMode) Validate() error {
	switch fm {
	case FileModeRegular, FileModeExec, FileModeDir, FileModeSymlink:
		return nil
	default:
		return fmt.Errorf("invalid file mode: %s", fm)
	}
}

// IsDir проверяет, является ли режим директорией
func (fm FileMode) IsDir() bool {
	return fm == FileModeDir
}
