// internal/commands/init_test.go
package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("Create basic repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := Init(tmpDir)
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		// Проверяем обязательные файлы
		required := []string{
			".sib/HEAD",
			".sib/objects",
			".sib/refs/heads",
			".sib/refs/tags",
		}

		for _, file := range required {
			path := filepath.Join(tmpDir, file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Required file/dir missing: %s", file)
			}
		}

		// Проверяем содержимое HEAD
		headPath := filepath.Join(tmpDir, ".sib", "HEAD")
		data, _ := os.ReadFile(headPath)
		if string(data) != "ref: refs/heads/master\n" {
			t.Errorf("HEAD contains wrong data: %s", string(data))
		}
	})
}
