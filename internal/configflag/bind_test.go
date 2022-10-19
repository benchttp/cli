package configflag_test

import (
	"flag"
	"reflect"
	"testing"

	"github.com/benchttp/engine/configparse"

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

		switch {
		case *repr.Request.Method != "POST",
			*repr.Request.URL != "https://benchttp.app?cool=yes",
			repr.Request.Header["API_KEY"][0] != "abc",
			repr.Request.Header["Accept"][0] != "text/html",
			repr.Request.Header["Accept"][1] != "application/json",
			repr.Request.Body.Type != "raw",
			repr.Request.Body.Content != "hello",

			*repr.Runner.Requests != 1,
			*repr.Runner.Concurrency != 2,
			*repr.Runner.Interval != "3s",
			*repr.Runner.RequestTimeout != "4s",
			*repr.Runner.GlobalTimeout != "5s":

			t.Error("unexpected parsed flags")
		}
	})
}
