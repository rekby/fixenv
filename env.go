package fixenv

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
)

var (
	globalCache               = newCache()
	globalEmptyFixtureOptions = &FixtureOptions{}
)

type EnvT struct {
	t T
	c *cache
}

func NewEnv(t T) *EnvT {
	return &EnvT{
		t: t,
		c: globalCache,
	}
}

func (e *EnvT) T() T {
	return e.t
}

func (e *EnvT) Cache(params interface{}, f FixtureInternalFunc, opt *FixtureOptions) interface{} {
	if opt == nil {
		opt = globalEmptyFixtureOptions
	}
	key, err := makeCacheKey(e.t.Name(), params, opt, false)
	if err != nil {
		e.t.Fatalf("failed to create cache key: %v", err)
		// return not reacheble after Fatalf
		return nil
	}
	wrappedF := fixtureCallWrapper(e.t, f, opt)
	res, err := e.c.GetOrSet(key, wrappedF)
	if err != nil {
		e.t.Fatalf("failed to call fixture func: %v", err)
		// return not reachable after Fatalf
		return nil
	}

	return res
}

// makeCacheKey generate cache key
// must be called from first level of env functions - for detect external caller
func makeCacheKey(testname string, params interface{}, opt *FixtureOptions, testCall bool) (cacheKey, error) {
	externalCallerLevel := 4
	var pc = make([]uintptr, externalCallerLevel)
	var extCallerFrame runtime.Frame
	if externalCallerLevel == runtime.Callers(0, pc) {
		frames := runtime.CallersFrames(pc)
		frames.Next()                     // callers
		frames.Next()                     // the function
		frames.Next()                     // caller of the function
		extCallerFrame, _ = frames.Next() // external caller
	}
	key := struct {
		Scope        CacheScope  `json:"scope"`
		ScopeName    string      `json:"scope_name"`
		FunctionName string      `json:"func"`
		FileName     string      `json:"fname"`
		LineNum      int         `json:"line"`
		Params       interface{} `json:"params"`
	}{
		Scope: opt.Scope,
		//ScopeName:    testname, // set above
		FunctionName: extCallerFrame.Function,
		FileName:     extCallerFrame.File,
		LineNum:      extCallerFrame.Line,
		Params:       params,
	}
	if testCall {
		key.LineNum = 123
		key.FileName = ".../" + filepath.Base(key.FileName)
	}

	switch opt.Scope {
	case ScopeTest:
		key.ScopeName = testname
	default:
		return "", fmt.Errorf("unexpected scope: %v", opt.Scope)
	}

	keyBytes, err := json.Marshal(key)
	if err != nil {
		return "", fmt.Errorf("failed to serialize params to json: %v", err)
	}
	return cacheKey(keyBytes), nil
}

func fixtureCallWrapper(t T, f FixtureInternalFunc, opt *FixtureOptions) FixtureInternalFunc {
	return func() (res interface{}, err error) {
		res, err = f()
		if opt.Scope == ScopeTest && opt.TearDown != nil {
			t.Cleanup(opt.TearDown)
		}
		return res, err
	}
}
