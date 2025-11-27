package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sib/internal/core/objects"
	"sib/internal/core/storage"
)

func main() {
	// Создаем временную директорию для демонстрации
	tmpDir := "./test-repo"
	os.RemoveAll(tmpDir)       // Очищаем перед началом
	defer os.RemoveAll(tmpDir) // Очищаем после завершения

	fmt.Println("🚀 Starting SibGit Demo...")
	fmt.Printf("📁 Using directory: %s\n\n", tmpDir)

	// Создаем хранилище
	store := storage.NewObjectStore(tmpDir)

	// 1. Демонстрация работы с Blob
	fmt.Println("1. 📄 Blob Operations:")

	// Создаем несколько тестовых файлов
	files := map[string]string{
		"README.md":   "# My Project\nThis is a test project",
		"main.go":     "package main\n\nfunc main() {\n    println(\"Hello, SibGit!\")\n}",
		"config.json": `{"name": "test", "version": "1.0.0"}`,
	}

	blobHashes := make(map[string]objects.Hash)

	for filename, content := range files {
		blob := objects.NewBlob([]byte(content))
		hash, err := store.WriteObject(blob)
		if err != nil {
			fmt.Printf("   ❌ Failed to write blob %s: %v\n", filename, err)
			continue
		}
		blobHashes[filename] = hash
		fmt.Printf("   ✅ %s -> %s\n", filename, hash)
	}

	// 2. Демонстрация работы с Tree
	fmt.Println("\n2. 📁 Tree Operations:")

	tree := objects.NewTree()

	for filename, hash := range blobHashes {
		entry, err := objects.NewTreeEntry(objects.FileModeRegular, filename, hash, objects.BlobObject)
		if err != nil {
			fmt.Printf("   ❌ Failed to create tree entry for %s: %v\n", filename, err)
			continue
		}
		tree.AddEntry(*entry)
		fmt.Printf("   ✅ Added to tree: %s\n", filename)
	}

	treeHash, err := store.WriteObject(tree)
	if err != nil {
		fmt.Printf("   ❌ Failed to write tree: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Tree saved: %s\n", treeHash)

	// Читаем tree обратно для проверки
	readTreeObj, err := store.ReadObject(treeHash)
	if err != nil {
		fmt.Printf("   ❌ Failed to read tree: %v\n", err)
	} else {
		readTree := readTreeObj.(*objects.Tree)
		fmt.Printf("   ✅ Tree entries: %d\n", len(readTree.Entries()))
	}

	// 3. Демонстрация работы с Commit
	fmt.Println("\n3. 🔖 Commit Operations:")

	author, _ := objects.NewSignature("Alex Developer", "alex@example.com", time.Now())
	committer, _ := objects.NewSignature("SibGit System", "sibgit@example.com", time.Now())

	commit, err := objects.NewCommit(treeHash, []objects.Hash{}, *author, *committer, "Initial commit with project structure")
	if err != nil {
		fmt.Printf("   ❌ Failed to create commit: %v\n", err)
		return
	}

	commitHash, err := store.WriteObject(commit)
	if err != nil {
		fmt.Printf("   ❌ Failed to write commit: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Commit saved: %s\n", commitHash)

	// Читаем commit обратно
	readCommitObj, err := store.ReadObject(commitHash)
	if err != nil {
		fmt.Printf("   ❌ Failed to read commit: %v\n", err)
	} else {
		readCommit := readCommitObj.(*objects.Commit)
		fmt.Printf("   ✅ Commit message: %s\n", readCommit.Message())
		author := readCommit.Author()
		fmt.Printf("   ✅ Commit author: %s <%s>\n", author.Name(), author.Email())
	}

	// 4. Демонстрация целостности данных
	fmt.Println("\n4. 🔒 Integrity Check:")

	// Проверяем существование объектов
	fmt.Println("   Checking object existence:")
	for _, hash := range blobHashes {
		if store.ObjectExists(hash) {
			fmt.Printf("   ✅ Object exists: %s\n", hash)
		} else {
			fmt.Printf("   ❌ Object missing: %s\n", hash)
		}
	}

	// 5. Показываем структуру хранилища
	fmt.Println("\n5. 🗂️ Storage Structure:")
	showStorageStructure(store, tmpDir)

	fmt.Println("\n🎉 Demo completed successfully!")
}

func showStorageStructure(store *storage.ObjectStore, basePath string) {
	objectsDir := filepath.Join(basePath, ".sib", "objects")

	fmt.Printf("   Storage path: %s\n", objectsDir)

	// Показываем поддиректории
	entries, err := os.ReadDir(objectsDir)
	if err != nil {
		fmt.Printf("   ❌ Cannot read objects directory: %v\n", err)
		return
	}

	for _, dir := range entries {
		if dir.IsDir() {
			subDirPath := filepath.Join(objectsDir, dir.Name())
			subEntries, _ := os.ReadDir(subDirPath)
			fmt.Printf("   📁 %s/ (%d objects)\n", dir.Name(), len(subEntries))

			// Показываем первые 3 объекта из каждой директории
			for i, file := range subEntries {
				if i < 3 {
					fmt.Printf("      📄 %s\n", file.Name())
				}
			}
			if len(subEntries) > 3 {
				fmt.Printf("      ... and %d more\n", len(subEntries)-3)
			}
		}
	}

	time.Sleep(time.Second * 300)
}
