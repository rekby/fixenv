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

func TestCache_DeleteKeys(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		at := assert.New(t)

		c := newCache()
		k1 := cacheKey("k1")
		k2 := cacheKey("k2")
		k3 := cacheKey("k3")
		val1 := "test1"
		valFunc := func() (interface{}, error) {
			return val1, nil
		}

		c.setOnce(k1, valFunc)
		c.setOnce(k2, valFunc)
		c.setOnce(k3, valFunc)

		c.DeleteKeys(k1, k2)
		_, ok := c.get(k1)
		at.False(ok)
		_, ok = c.get(k2)
		at.False(ok)
		res, ok := c.get(k3)
		at.True(ok)
		at.Equal(val1, res.res)

		val2 := "test2"
		c.setOnce(k1, func() (res interface{}, err error) {
			return val2, nil
		})
		res, ok = c.get(k1)
		at.True(ok)
		at.Equal(val2, res.res)
	})

	t.Run("mutex", func(t *testing.T) {
		at := assert.New(t)

		c := newCache()
		c.setOnce("asd", func() (res interface{}, err error) {
			return nil, nil
		})

		c.m.RLock()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DeleteKeys("asd")
			wg.Done()
		}()

		time.Sleep(waitTime)
		at.Len(c.store, 1)
		at.Len(c.setLocks, 1)

		c.m.RUnlock()
		wg.Wait()
		at.Len(c.store, 0)
		at.Len(c.setLocks, 0)
	})
}

func TestCache_Get(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()

		_, ok := c.get("qwe")
		at.False(ok)

		c.store["asd"] = cacheVal{res: "val"}

		res, ok := c.get("asd")
		at.True(ok)
		at.Equal(cacheVal{res: "val"}, res)
	})

	t.Run("read_mutex", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		c.setOnce("asd", func() (res interface{}, err error) {
			return nil, nil
		})
		c.m.RLock()
		_, ok := c.get("asd")
		c.m.RUnlock()
		at.True(ok)
	})

	t.Run("write_mutex", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		c.setOnce("asd", func() (res interface{}, err error) {
			return nil, nil
		})
		c.m.Lock()
		var ok bool
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			_, ok = c.get("asd")
			wg.Done()
		}()

		time.Sleep(waitTime)
		at.False(ok)

		c.m.Unlock()
		wg.Wait()

		at.True(ok)
	})

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

		var firstMuStarted = make(chan bool)

		var firstMuNeedFinish = make(chan bool)

		var firstMuFinished = make(chan bool)

		// first
		go func() {
			c.setOnce(key, func() (res interface{}, err error) {
				close(firstMuStarted)
				<-firstMuNeedFinish

				return 1, nil
			})

			close(firstMuFinished)
		}()

		// wait first func start execute
		<-firstMuStarted

		var secondFinished = make(chan bool)

		var doneSecond int64
		go func() {
			c.setOnce(key, func() (res interface{}, err error) {
				// func will not call never
				return 2, nil
			})

			// must executed only after fist func finished
			atomic.AddInt64(&doneSecond, 1)
			close(secondFinished)
		}()

		time.Sleep(waitTime)

		// second not work until first finished
		doneSecondVal := atomic.LoadInt64(&doneSecond)
		_, ok := c.store[key]
		at.False(ok)
		at.Equal(int64(0), doneSecondVal)

		close(firstMuNeedFinish)

		// wait first func finished
		<-firstMuFinished

		// wait second finished
		<-secondFinished

		doneSecondVal = atomic.LoadInt64(&doneSecond)
		at.Equal(1, c.store[key].res)
		at.Equal(int64(1), doneSecondVal)
	})

	t.Run("second_func_other_key_work", func(t *testing.T) {
		at := assert.New(t)
		c := newCache()
		key1 := cacheKey("1")
		key2 := cacheKey("2")

		var firstMuStarted = make(chan bool)

		var firstMuNeedFinish = make(chan bool)

		var firstMuFinished = make(chan bool)

		// first
		go func() {
			c.setOnce(key1, func() (res interface{}, err error) {
				close(firstMuStarted)

				<-firstMuNeedFinish

				return 1, nil
			})

			close(firstMuFinished)
		}()

		// wait first func start execute
		<-firstMuStarted

		// call second func in same goroutine
		c.setOnce(key2, func() (res interface{}, err error) {
			// func will not call never
			return 2, nil
		})

		// allow finish first func after second already finished
		close(firstMuNeedFinish)

		// wait first func finished
		<-firstMuFinished

		at.Equal(1, c.store[key1].res)
		at.Equal(2, c.store[key2].res)
	})
}

func TestCache_GetOrSetRaceCondition(_ *testing.T) {
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
