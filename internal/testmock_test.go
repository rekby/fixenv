package internal

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTestMock_CallCleanup(t *testing.T) {
	tm := &TestMock{}
	called := false
	tm.Cleanup(func() {
		called = true
	})
	tm.CallCleanup()
	if !called {
		t.Fatal()
	}
}

func TestTestMock_Fatalf(t *testing.T) {
	t.Run("SkipExit", func(t *testing.T) {
		tm := &TestMock{
			SkipGoexit: true,
		}
		tm.Fatalf("a: %v", 123)
		if !reflect.DeepEqual(tm.Fatals, []FormatCall{
			{
				Format:       "a: %v",
				Args:         []interface{}{123},
				ResultString: fmt.Sprintf("a: %v", 123),
			},
		}) {
			t.Fatal(tm.Fatals)
		}
	})
	t.Run("WithExit", func(t *testing.T) {
		tm := &TestMock{}
		called := false
		exit := make(chan bool)
		go func() {
			defer close(exit)
			tm.Fatalf("a: %v", 123)
			called = true
		}()
		<-exit
		if !reflect.DeepEqual(tm.Fatals, []FormatCall{
			{
				Format:       "a: %v",
				Args:         []interface{}{123},
				ResultString: fmt.Sprintf("a: %v", 123),
			},
		}) {
			t.Fatal(tm.Fatals)
		}
		if called {
			t.Fatal()
		}
	})
}

func TestTestMock_Logf(t *testing.T) {
	tm := &TestMock{}
	tm.Logf("a: %v", 123)
	if !reflect.DeepEqual(tm.Logs, []FormatCall{
		{
			Format:       "a: %v",
			Args:         []interface{}{123},
			ResultString: fmt.Sprintf("a: %v", 123),
		},
	}) {
		t.Fatal(tm.Logs)
	}
}

func TestTestMock_Name(t *testing.T) {
	tm := &TestMock{}
	if tm.Name() != defaultMockTestName {
		t.Fatal(tm.Name())
	}

	tm.TestName = "qwe"
	if tm.Name() != "qwe" {
		t.Fatal(tm.Name())
	}
}

func TestTestMock_SkipNow(t *testing.T) {
	t.Run("SkipExit", func(t *testing.T) {
		tm := &TestMock{SkipGoexit: true}
		tm.SkipNow()
		if tm.SkipCount != 1 {
			t.Fatal(tm.SkipCount)
		}
		tm.SkipNow()
		if tm.SkipCount != 2 {
			t.Fatal(tm.SkipCount)
		}
	})
	t.Run("WithExit", func(t *testing.T) {
		tm := &TestMock{}
		exit := make(chan bool)

		go func() {
			defer close(exit)
			tm.SkipNow()
		}()
		<-exit
		if tm.SkipCount != 1 {
			t.Fatal(tm.SkipCount)
		}
		exit = make(chan bool)

		go func() {
			defer close(exit)
			tm.SkipNow()
		}()
		<-exit
		if tm.SkipCount != 2 {
			t.Fatal(tm.SkipCount)
		}
	})
}

func TestTestMock_Skipped(t *testing.T) {
	tm := &TestMock{}
	if tm.Skipped() {
		t.Fatal()
	}

	tm.SkipCount = 1
	if !tm.Skipped() {
		t.Fatal()
	}

	tm.SkipCount = 2
	if !tm.Skipped() {
		t.Fatal()
	}
}
