package sf

import (
	"errors"
	"github.com/rekby/fixenv"
	"github.com/rekby/fixenv/internal"
	"testing"
)

func TestMustNoError(t *testing.T) {
	tm := &internal.TestMock{}
	t.Cleanup(tm.CallCleanup)

	tm.SkipGoexit = true
	e := fixenv.New(tm)
	mustNoErr(e, nil, "a: %v", 123)

	if len(tm.Fatals) > 0 {
		t.Fatalf("no err must be simple return: %#v", tm.Fatals)
	}

	mustNoErr(e, errors.New("asd"), "b: %v", 123)
	if len(tm.Fatals) != 1 {
		t.Fatalf("fatals: %#v", tm.Fatals)
	}
	if tm.Fatals[0].Format != "b: %v" {
		t.Fatal(tm.Fatals[0].Format)
	}
	if tm.Fatals[0].Args[0] != 123 {
		t.Fatal(tm.Fatals[0].Args[0])
	}
}
