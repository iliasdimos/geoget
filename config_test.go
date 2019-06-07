package main

import (
	"os"
	"testing"
)

func Test_newCfg(t *testing.T) {

	name := "Set log level"
	env := "LOG_LEVEL"
	set := "debug"
	want := "debug"

	err := os.Setenv(env, set)
	if err != nil {
		t.Errorf("%s error setting env variable", name)
	}
	got, err := newCfg("")
	if err != nil {
		t.Errorf("%s error reading config from env variable", name)
	}
	if got.Log.Level != want {
		t.Errorf("%s: newCfg().Log.Level = %v, want %v", name, got.Log.Level, want)
	}
}
