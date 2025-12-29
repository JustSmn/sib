// internal/commands/add_test.go
package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	t.Run("Add all files in repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 1. Init
		err := Init(tmpDir)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		// 2. Create test files
		files := map[string]string{
			"main.go":     "package main",
			"README.md":   "# Test",
			"src/util.go": "// util",
		}

		for path, content := range files {
			fullPath := filepath.Join(tmpDir, path)
			os.MkdirAll(filepath.Dir(fullPath), 0755)
			os.WriteFile(fullPath, []byte(content), 0644)
		}

		// 3. Add
		err = Add(tmpDir)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// 4. Check index was created
		indexPath := filepath.Join(tmpDir, ".sib", "index")
		info, err := os.Stat(indexPath)
		if err != nil {
			t.Fatalf("Index not created: %v", err)
		}

		if info.Size() == 0 {
			t.Error("Index file is empty")
		}
	})

	t.Run("Add fails outside repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := Add(tmpDir)
		if err == nil {
			t.Error("Expected error when adding outside repo")
		}
	})
}
