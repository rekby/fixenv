package simple

import (
	"github.com/rekby/fixenv"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(fixenv.RunTests(m))
}
