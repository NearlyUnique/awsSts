package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\texp: %#v\n\tgot: %#v\033[39m\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func mapFromString(tb testing.TB, body string) (o map[string]*json.RawMessage) {
	err := json.Unmarshal([]byte(body), &o)
	ok(tb, err)
	return o
}

func mapFromBytes(tb testing.TB, body []byte) (o map[string]*json.RawMessage) {
	err := json.Unmarshal(body, &o)
	ok(tb, err)
	return o
}
func isSubSet(super, sub []string) bool {
	if len(super) == 0 && len(sub) == 0 {
		return true
	}
	if len(sub) == 0 {
		return false
	}
	cpy := append([]string{}, sub...)
	for _, a := range super {
		for i, b := range cpy {
			if a == b {
				cpy = append(cpy[:i], cpy[i+1:]...)
				break
			}
		}
	}
	if len(cpy) == 0 {
		return true
	}
	return false
}
