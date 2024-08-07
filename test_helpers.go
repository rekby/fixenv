package fixenv

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func requireTrue(t *testing.T, val bool) {
	t.Helper()
	if !val {
		t.Fatal("Must be true")
	}
}

func requireFalse(t *testing.T, val bool) {
	t.Helper()
	if val {
		t.Fatal("Must be false")
	}
}

func requireEquals(t *testing.T, v1, v2 interface{}) {
	t.Helper()

	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("Must be equals")
	}
}

func requireNotEquals(t *testing.T, v1, v2 interface{}) {
	if reflect.DeepEqual(v1, v2) {
		t.Fatal("Must be not equals")
	}

}

func isError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Must be error")
	}
}
func noError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("err must be nil, but got: %+v", err)
	}
}

func requireNil(t *testing.T, v interface{}) {
	t.Helper()
	if v == nil {
		return
	}

	val := reflect.ValueOf(v)
	if !val.IsNil() {
		t.Fatal("Must be nil")
	}
}

func requireNotNil(t *testing.T, v interface{}) {
	t.Helper()
	if v == nil {
		return
	}

	val := reflect.ValueOf(v)
	if val.IsNil() {
		t.Fatal("Must be not nil")
	}
}

func requireContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("'%s' must contains '%s'", s, substr)
	}
}

func requirePanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		_ = recover()
	}()

	f()
	t.Fatal("the function must raise panic")
}

func requireJSONEquals(t *testing.T, v1 string, v2 string) {
	t.Helper()
	var vj1 interface{}
	var vj2 interface{}

	err := json.Unmarshal([]byte(v1), &vj1)
	noError(t, err)

	err = json.Unmarshal([]byte(v2), &vj2)
	noError(t, err)

	data1, err := json.Marshal(vj1)
	noError(t, err)

	data2, err := json.Marshal(vj2)
	noError(t, err)

	normalized1 := string(data1)
	normalized2 := string(data2)

	requireEquals(t, normalized1, normalized2)
}
