package fixenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeCacheKey(t *testing.T) {
	at := assert.New(t)

	var res cacheKey
	var err error

	envFunc := func() {
		res, err = makeCacheKey("asdf", 222, globalEmptyFixtureOptions, true)
	}
	envFunc()
	at.NoError(err)

	expected := cacheKey(`{"func":"fixenv.TestMakeCacheKey","fname":".../env_test.go","line":123,"scope":0,"scope_name":"asdf","params":222}`)
	at.JSONEq(string(expected), string(res))
}
