package sf

import (
	"github.com/rekby/fixenv"
	"os"
	"testing"
)

func TestTempDir(t *testing.T) {
	var dir string
	t.Run("subtest", func(t *testing.T) {
		e := fixenv.New(t)
		dir = TempDir(e)
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("failed check tmp dir: %v", err)
		}
	})
	_, err := os.Stat(dir)
	if !os.IsNotExist(err) {
		t.Fatalf("Directory must be removed after test finished, have err: %v", err)
	}
}

func TestTempFile(t *testing.T) {
	var file string
	t.Run("subtest", func(t *testing.T) {
		e := fixenv.New(t)
		file := TempFile(e)
		if _, err := os.Stat(file); err != nil {
			t.Fatal(err)
		}
	})

	_, err := os.Stat(file)
	if !os.IsNotExist(err) {
		t.Fatal("File must be removed with temp directory")
	}
}
