package configflag_test

import (
	"flag"
	"reflect"
	"testing"

	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/configflag"
)

func TestWhich(t *testing.T) {
	for _, tc := range []struct {
		label string
		args  []string
		exp   []string
	}{
		{
			label: "return all config flags set",
			args: []string{
				"-method", "POST",
				"-url", "https://benchttp.app?cool=yes",
				"-concurrency", "2",
				"-requests", "3",
				"-requestTimeout", "1s",
				"-globalTimeout", "4s",
			},
			exp: []string{
				"concurrency", "globalTimeout", "method",
				"requestTimeout", "requests", "url",
			},
		},
		{
			label: "do not return config flags not set",
			args:  []string{"-requests", "3"},
			exp:   []string{"requests"},
		},
	} {
		flagset := flag.NewFlagSet("run", flag.ExitOnError)

		configflag.Bind(flagset, &runner.Config{})

		if err := flagset.Parse(tc.args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		if got := configflag.Which(flagset); !reflect.DeepEqual(got, tc.exp) {
			t.Errorf("\nexp %v\ngot %v", tc.exp, got)
		}
	}
}
