package fixenv

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const waitTime = time.Second / 10

func TestCache_Get(t *testing.T) {
	at := assert.New(t)

	c := newCache()

	_, ok := c.get("qwe")
	at.False(ok)

	c.store["asd"] = cacheVal{res: "val"}

	res, ok := c.get("asd")
	at.True(ok)
	at.Equal(cacheVal{res: "val"}, res)
}

func TestCache_SetOnce(t *testing.T) {
	t.Run("save_new_key", func(t *testing.T) {
		at := assert.New(t)

		c := newCache()
		cnt := 0
		key1 := cacheKey("1")
		c.setOnce(key1, func() (res interface{}, err error) {
			cnt++
			return 1, nil
		})
		at.Equal(1, cnt)
		at.Equal(1, c.store[key1].res)
		at.NoError(c.store[key1].err)

		c.setOnce(key1, func() (res interface{}, err error) {
			cnt++
			return 2, nil
		})
		at.Equal(1, cnt)
		at.Equal(1, c.store[key1].res)
		at.NoError(c.store[key1].err)
	})

	t.Run("second_set_val", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		key1 := cacheKey("1")
		key2 := cacheKey("2")
		cnt := 0
		c.setOnce(key1, func() (res interface{}, err error) {
			cnt++
			return 1, nil
		})
		c.setOnce(key2, func() (res interface{}, err error) {
			cnt++
			return 2, nil
		})
		at.Equal(2, cnt)
		at.Equal(1, c.store[key1].res)
		at.NoError(c.store[key1].err)
		at.Equal(2, c.store[key2].res)
		at.NoError(c.store[key2].err)
	})

	// exit without return value
	t.Run("exit_without_return", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		key := cacheKey("3")
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.setOnce(key, func() (res interface{}, err error) {
				runtime.Goexit()
				return 3, nil
			})
		}()
		wg.Wait()

		at.Nil(c.store[key].res)
		at.Error(c.store[key].err)
	})

	t.Run("second_func_same_key_wait", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		key := cacheKey("1")

		var firstMuStarted sync.Mutex
		firstMuStarted.Lock()

		var firstMuNeedFinish sync.Mutex
		firstMuNeedFinish.Lock()

		var firstMuFinished sync.Mutex
		firstMuFinished.Lock()

		// first
		go func() {
			c.setOnce(key, func() (res interface{}, err error) {
				firstMuStarted.Unlock()

				firstMuNeedFinish.Lock()
				defer firstMuNeedFinish.Unlock()

				return 1, nil
			})

			firstMuFinished.Unlock()
		}()

		// wait first func start execute
		firstMuStarted.Lock()
		firstMuStarted.Unlock()

		var doneSecond int64
		go func() {
			c.setOnce(key, func() (res interface{}, err error) {
				// func will not call never
				return 2, nil
			})

			// must executed only after fist func finished
			atomic.AddInt64(&doneSecond, 1)
		}()

		time.Sleep(waitTime)
		doneSecondVal := atomic.LoadInt64(&doneSecond)
		_, ok := c.store[key]
		at.False(ok)
		at.Equal(int64(0), doneSecondVal)

		firstMuNeedFinish.Unlock()

		// wait first func finished
		firstMuFinished.Lock()
		firstMuFinished.Unlock()

		// give time for execute second func
		time.Sleep(waitTime)

		doneSecondVal = atomic.LoadInt64(&doneSecond)
		at.Equal(1, c.store[key].res)
		at.Equal(int64(1), doneSecondVal)
	})

	t.Run("second_func_other_key_work", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		key1 := cacheKey("1")
		key2 := cacheKey("2")

		var firstMuStarted sync.Mutex
		firstMuStarted.Lock()

		var firstMuNeedFinish sync.Mutex
		firstMuNeedFinish.Lock()

		var firstMuFinished sync.Mutex
		firstMuFinished.Lock()

		// first
		go func() {
			c.setOnce(key1, func() (res interface{}, err error) {
				firstMuStarted.Unlock()

				firstMuNeedFinish.Lock()
				defer firstMuNeedFinish.Unlock()

				return 1, nil
			})

			firstMuFinished.Unlock()
		}()

		// wait first func start execute
		firstMuStarted.Lock()
		firstMuStarted.Unlock()

		// call second func in same goroutine
		c.setOnce(key2, func() (res interface{}, err error) {
			// func will not call never
			return 2, nil
		})

		// allow finish first func after second already finished
		firstMuNeedFinish.Unlock()

		// wait first func finished
		firstMuFinished.Lock()
		firstMuFinished.Unlock()
		at.Equal(1, c.store[key1].res)
		at.Equal(2, c.store[key2].res)
	})
}

func TestCache_GetOrSetRaceCondition(t *testing.T) {
	parallels := 100
	iterations := 1000

	rndMaxBound := 1000

	c := newCache()

	var wg sync.WaitGroup
	for i := 0; i < parallels; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				key := cacheKey(strconv.Itoa(rand.Intn(rndMaxBound)))
				v, ok := c.GetOrSet(key, func() (res interface{}, err error) {
					return 1, nil
				})
				_ = v
				_ = ok
			}
		}()
	}
	wg.Wait()
}
