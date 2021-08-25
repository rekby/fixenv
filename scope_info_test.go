package fixenv

import (
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScopeInfo_AddKey(t *testing.T) {
	at := assert.New(t)

	t.Run("simple", func(t *testing.T) {
		si := newScopeInfo(t)
		at.Equal(t, si.t)
		at.Len(si.cacheKeys, 0)

		si.AddKey("asd")
		si.AddKey("ddd")
		at.Equal([]cacheKey{"asd", "ddd"}, si.cacheKeys)
	})

	t.Run("race", func(t *testing.T) {
		si := newScopeInfo(t)

		count := 10000
		source := make([]cacheKey, count)
		for i := 0; i < count; i++ {
			source[i] = cacheKey(strconv.Itoa(i))
		}

		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(key cacheKey) {
				si.AddKey(key)
				wg.Done()
			}(source[i])
		}
		wg.Wait()

		sort.Slice(si.cacheKeys, func(i, j int) bool {
			iInt, _ := strconv.Atoi(string(si.cacheKeys[i]))
			jInt, _ := strconv.Atoi(string(si.cacheKeys[j]))

			return iInt < jInt
		})

		at := assert.New(t)
		at.Equal(source, si.cacheKeys)
	})
}

func TestScopeInfo_Keys(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		at := assert.New(t)

		si := newScopeInfo(t)
		si.AddKey("asd")
		si.AddKey("kkk")

		keys := si.Keys()
		at.Equal([]cacheKey{"asd", "kkk"}, keys)
	})

	t.Run("mutex", func(t *testing.T) {
		at := assert.New(t)

		si := newScopeInfo(t)
		si.AddKey("asd")
		si.AddKey("kkk")

		si.m.Lock()
		var keys []cacheKey
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			keys = si.Keys()
			wg.Done()
		}()

		time.Sleep(waitTime)
		at.Len(keys, 0)

		si.m.Unlock()
		wg.Wait()
		at.Equal([]cacheKey{"asd", "kkk"}, keys)
	})
}
