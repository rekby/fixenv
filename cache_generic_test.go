package fixenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrSet(t *testing.T) {
	at := assert.New(t)

	e := NewEnv(t)

	res := GetOrSet(e, nil, nil, func() (res int, err error) {
		return 1, nil
	})

	if res+1 != 2 {
		// check result and compile res + 1 - to check if res is int for compiler
		at.Fail("1+1 != 2")
	}
	at.Equal(1, res)
}
