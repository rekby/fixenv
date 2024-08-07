package fixenv

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const waitTime = time.Second / 10

func TestCache_DeleteKeys(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		c := newCache()
		k1 := cacheKey("k1")
		k2 := cacheKey("k2")
		k3 := cacheKey("k3")
		val1 := "test1"
		valFunc := func() (*Result, error) {
			return NewResult(val1), nil
		}

		c.setOnce(k1, valFunc)
		c.setOnce(k2, valFunc)
		c.setOnce(k3, valFunc)

		c.DeleteKeys(k1, k2)
		_, ok := c.get(k1)
		requireFalse(t, ok)
		_, ok = c.get(k2)
		requireFalse(t, ok)
		res, ok := c.get(k3)
		requireTrue(t, ok)
		requireEquals(t, val1, res.res.Value.(string))

		val2 := "test2"
		c.setOnce(k1, func() (res *Result, err error) {
			return NewResult(val2), nil
		})
		res, ok = c.get(k1)
		requireTrue(t, ok)
		requireEquals(t, val2, res.res.Value.(string))
	})

	t.Run("mutex", func(t *testing.T) {
		c := newCache()
		c.setOnce("asd", func() (res *Result, err error) {
			return NewResult(nil), nil
		})

		c.m.RLock()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			c.DeleteKeys("asd")
			wg.Done()
		}()

		time.Sleep(waitTime)
		requireEquals(t, len(c.store), 1)
		requireEquals(t, len(c.setLocks), 1)

		c.m.RUnlock()
		wg.Wait()
		requireEquals(t, len(c.store), 0)
		requireEquals(t, len(c.setLocks), 0)
	})
}

func TestCache_Get(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		c := newCache()

		_, ok := c.get("qwe")
		requireFalse(t, ok)

		c.store["asd"] = cacheVal{res: NewResult("val")}

		res, ok := c.get("asd")
		requireTrue(t, ok)
		requireEquals(t, cacheVal{res: NewResult("val")}, res)
	})

	t.Run("read_mutex", func(t *testing.T) {
		c := newCache()
		c.setOnce("asd", func() (res *Result, err error) {
			return NewResult(nil), nil
		})
		c.m.RLock()
		_, ok := c.get("asd")
		c.m.RUnlock()
		requireTrue(t, ok)
	})

	t.Run("write_mutex", func(t *testing.T) {
		c := newCache()
		c.setOnce("asd", func() (res *Result, err error) {
			return NewResult(nil), nil
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
		requireFalse(t, ok)

		c.m.Unlock()
		wg.Wait()

		requireTrue(t, ok)
	})

}

func TestCache_SetOnce(t *testing.T) {
	t.Run("save_new_key", func(t *testing.T) {
		c := newCache()
		cnt := 0
		key1 := cacheKey("1")
		c.setOnce(key1, func() (res *Result, err error) {
			cnt++
			return NewResult(1), nil
		})
		requireEquals(t, 1, cnt)
		requireEquals(t, 1, c.store[key1].res.Value)
		noError(t, c.store[key1].err)

		c.setOnce(key1, func() (res *Result, err error) {
			cnt++
			return NewResult(2), nil
		})
		requireEquals(t, 1, cnt)
		requireEquals(t, 1, c.store[key1].res.Value)
		noError(t, c.store[key1].err)
	})

	t.Run("second_set_val", func(t *testing.T) {
		c := newCache()
		key1 := cacheKey("1")
		key2 := cacheKey("2")
		cnt := 0
		c.setOnce(key1, func() (res *Result, err error) {
			cnt++
			return NewResult(1), nil
		})
		c.setOnce(key2, func() (res *Result, err error) {
			cnt++
			return NewResult(2), nil
		})
		requireEquals(t, 2, cnt)
		requireEquals(t, 1, c.store[key1].res.Value)
		noError(t, c.store[key1].err)
		requireEquals(t, 2, c.store[key2].res.Value)
		noError(t, c.store[key2].err)
	})

	// exit without return value
	t.Run("exit_without_return", func(t *testing.T) {
		c := newCache()
		key := cacheKey("3")
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.setOnce(key, func() (res *Result, err error) {
				runtime.Goexit()
				return NewResult(3), nil
			})
		}()
		wg.Wait()

		requireNil(t, c.store[key].res)
		isError(t, c.store[key].err)
	})

	t.Run("second_func_same_key_wait", func(t *testing.T) {
		c := newCache()
		key := cacheKey("1")

		var firstMuStarted = make(chan bool)

		var firstMuNeedFinish = make(chan bool)

		var firstMuFinished = make(chan bool)

		// first
		go func() {
			c.setOnce(key, func() (res *Result, err error) {
				close(firstMuStarted)
				<-firstMuNeedFinish

				return NewResult(1), nil
			})

			close(firstMuFinished)
		}()

		// wait first func start execute
		<-firstMuStarted

		var secondFinished = make(chan bool)

		var doneSecond int64
		go func() {
			c.setOnce(key, func() (res *Result, err error) {
				// func will not call never
				return NewResult(2), nil
			})

			// must executed only after fist func finished
			atomic.AddInt64(&doneSecond, 1)
			close(secondFinished)
		}()

		time.Sleep(waitTime)

		// second not work until first finished
		doneSecondVal := atomic.LoadInt64(&doneSecond)
		_, ok := c.store[key]
		requireFalse(t, ok)
		requireEquals(t, int64(0), doneSecondVal)

		close(firstMuNeedFinish)

		// wait first func finished
		<-firstMuFinished

		// wait second finished
		<-secondFinished

		doneSecondVal = atomic.LoadInt64(&doneSecond)
		requireEquals(t, 1, c.store[key].res.Value)
		requireEquals(t, int64(1), doneSecondVal)
	})

	t.Run("second_func_other_key_work", func(t *testing.T) {
		c := newCache()
		key1 := cacheKey("1")
		key2 := cacheKey("2")

		var firstMuStarted = make(chan bool)

		var firstMuNeedFinish = make(chan bool)

		var firstMuFinished = make(chan bool)

		// first
		go func() {
			c.setOnce(key1, func() (res *Result, err error) {
				close(firstMuStarted)

				<-firstMuNeedFinish

				return NewResult(1), nil
			})

			close(firstMuFinished)
		}()

		// wait first func start execute
		<-firstMuStarted

		// call second func in same goroutine
		c.setOnce(key2, func() (res *Result, err error) {
			// func will not call never
			return NewResult(2), nil
		})

		// allow finish first func after second already finished
		close(firstMuNeedFinish)

		// wait first func finished
		<-firstMuFinished

		requireEquals(t, 1, c.store[key1].res.Value)
		requireEquals(t, 2, c.store[key2].res.Value)
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
				v, ok := c.GetOrSet(key, func() (res *Result, err error) {
					return NewResult(1), nil
				})
				_ = v
				_ = ok
			}
		}()
	}
	wg.Wait()
}
