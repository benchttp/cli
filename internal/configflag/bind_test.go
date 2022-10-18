package configflag_test

import (
	"flag"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/engine/configparse"
	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/configflag"
)

func TestBind(t *testing.T) {
	t.Run("default to zero representation", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ExitOnError)
		args := []string{} // no args

		repr := configparse.Representation{}
		configflag.Bind(flagset, &repr)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		exp := configparse.Representation{}
		if got := repr; !reflect.DeepEqual(got, exp) {
			t.Errorf("\nexp %#v\ngot %#v", exp, got)
		}
	})

	t.Run("set config with flags values", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ExitOnError)
		args := []string{
			"-method", "POST",
			"-url", "https://benchttp.app?cool=yes",
			"-header", "API_KEY:abc",
			"-header", "Accept:text/html",
			"-header", "Accept:application/json",
			"-body", "raw:hello",
			"-requests", "1",
			"-concurrency", "2",
			"-interval", "3s",
			"-requestTimeout", "4s",
			"-globalTimeout", "5s",
		}

		repr := configparse.Representation{}
		configflag.Bind(flagset, &repr)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		exp := runner.Config{
			Request: runner.RequestConfig{
				Method: "POST",
				Header: http.Header{
					"API_KEY": {"abc"},
					"Accept":  {"text/html", "application/json"},
				},
				Body: runner.RequestBody{Type: "raw", Content: []byte("hello")},
			}.WithURL("https://benchttp.app?cool=yes"),
			Runner: runner.RecorderConfig{
				Requests:       1,
				Concurrency:    2,
				Interval:       3 * time.Second,
				RequestTimeout: 4 * time.Second,
				GlobalTimeout:  5 * time.Second,
			},
		}

		var got runner.Config
		if err := repr.Unmarshal(&got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, exp) {
			t.Errorf("\nexp %#v\ngot %#v", exp, got)
		}
	})
}
