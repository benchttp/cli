package configflag_test

import (
	"flag"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/configflag"
)

func TestBind(t *testing.T) {
	t.Run("default to base config", func(t *testing.T) {
		flagset := flag.NewFlagSet("run", flag.ExitOnError)
		args := []string{} // no args

		cfg := runner.DefaultConfig()
		configflag.Bind(flagset, &cfg)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		if exp := runner.DefaultConfig(); !reflect.DeepEqual(cfg, exp) {
			t.Errorf("\nexp %#v\ngot %#v", exp, cfg)
		}
	})

	t.Run("set config with flags values", func(t *testing.T) {
		flagset := flag.NewFlagSet("run", flag.ExitOnError)
		args := []string{
			"-method", "POST",
			"-url", "https://benchttp.app?cool=yes",
			"-header", "Content-Type:application/json",
			"-body", "raw:hello",
			"-requests", "1",
			"-concurrency", "2",
			"-interval", "3s",
			"-requestTimeout", "4s",
			"-globalTimeout", "5s",
		}

		cfg := runner.Config{}
		configflag.Bind(flagset, &cfg)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		exp := runner.Config{
			Request: runner.RequestConfig{
				Method: "POST",
				Header: http.Header{"Content-Type": {"application/json"}},
				Body:   runner.RequestBody{Type: "raw", Content: []byte("hello")},
			}.WithURL("https://benchttp.app?cool=yes"),
			Runner: runner.RecorderConfig{
				Requests:       1,
				Concurrency:    2,
				Interval:       3 * time.Second,
				RequestTimeout: 4 * time.Second,
				GlobalTimeout:  5 * time.Second,
			},
		}

		if !reflect.DeepEqual(cfg, exp) {
			t.Errorf("\nexp %#v\ngot %#v", exp, cfg)
		}
	})
}
