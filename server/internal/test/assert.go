package test

import (
	"reflect"
	"testing"
)

func AssertNoError(t *testing.T, err error, s ...string) {
	if err != nil {
		if len(s) != 0 {
			t.Fatalf(s[0], err)
		} else {
			t.Fatal(err)
		}
	}
}

func AssertEqual(t *testing.T, want interface{}, got interface{}) {
	tw := reflect.TypeOf(want)
	tg := reflect.TypeOf(got)
	if tw != tg {
		t.Errorf("Want and got are different types\nwant: %s, got: %s",
			tw.Name(), tg.Name(),
		)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Want and got are not equal:\nwant: %+v\ngot:%+v",
			want, got,
		)
	}
}
