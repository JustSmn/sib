package objects

// Serializable определяет контракт для сериализации объектов
// Используется CAS-хранилищем для единообразной работы со всеми типами объектов
// "Любой объект, который можно сериализовать, должен иметь эти методы"
type Serializable interface {
	Serialize() ([]byte, error) // Serialize преобразует объект в байтовое представление для хеширования и сохранения
	Type() ObjectType           // Type возвращает тип объекта (blob, tree, commit, tag)
}

// Hashable определяет контракт для работы с хешами
// Позволяет единообразно работать с хешами разных объектов
// "Любой объект с хешом должен уметь его устанавливать и возвращать"
type Hashable interface {
	SetHash(Hash)  // SetHash сохраняет вычисленный хеш в объекте
	GetHash() Hash // GetHash возвращает хеш объекта
}

/*Зачем это нужно:
go
// Без интерфейсов в storage пришлось бы делать так:
func WriteObject(obj interface{}) error {
    switch v := obj.(type) {
    case *Blob:
        data, err := v.Serialize()
    case *Tree:
        data, err := v.Serialize()
    case *Commit:
        // ... и так для каждого типа ❌
    }
}

// С интерфейсами становится ЕДИНЫМ:
func WriteObject(obj Serializable) error {
    data, err := obj.Serialize()  // ✅ Работает для ЛЮБОГО объекта!
    // ...
}*/
