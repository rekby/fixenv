package internal

import (
	"fmt"
	"runtime"
	"sync"
)

const defaultMockTestName = "mock"

type FormatCall struct {
	Format       string
	Args         []interface{}
	ResultString string
}

type TestMock struct {
	TestName   string
	SkipGoexit bool

	M         sync.Mutex
	Cleanups  []func()
	Logs      []FormatCall
	Fatals    []FormatCall
	SkipCount int
}

func (t *TestMock) CallCleanup() {
	for i := len(t.Cleanups) - 1; i >= 0; i-- {
		t.Cleanups[i]()
	}
}

func (t *TestMock) Cleanup(f func()) {
	t.M.Lock()
	defer t.M.Unlock()

	t.Cleanups = append(t.Cleanups, f)
}

func (t *TestMock) Fatalf(format string, args ...interface{}) {
	t.M.Lock()
	defer t.M.Unlock()

	t.Fatals = append(t.Fatals, FormatCall{
		Format:       format,
		Args:         args,
		ResultString: fmt.Sprintf(format, args...),
	})

	if !t.SkipGoexit {
		runtime.Goexit()
	}
}

func (t *TestMock) Logf(format string, args ...interface{}) {
	t.M.Lock()
	defer t.M.Unlock()

	t.Logs = append(t.Logs, FormatCall{Format: format, Args: args, ResultString: fmt.Sprintf(format, args...)})
}

func (t *TestMock) Name() string {
	if t.TestName == "" {
		return defaultMockTestName
	}
	return t.TestName
}

func (t *TestMock) SkipNow() {
	t.M.Lock()
	t.SkipCount++
	t.M.Unlock()

	if !t.SkipGoexit {
		runtime.Goexit()
	}
}

func (t *TestMock) Skipped() bool {
	t.M.Lock()
	defer t.M.Unlock()

	return t.SkipCount > 0
}
