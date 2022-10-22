package configfile_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/configfile"
	"github.com/benchttp/cli/internal/testutil"
)

const (
	validConfigPath = "./testdata"
	validURL        = "http://localhost:9999?fib=30&delay=200ms" // value from testdata files
)

var supportedExt = []string{
	".yml",
	".yaml",
	".json",
}

// TestParse ensures the config file is open, read, and correctly parsed.
func TestParse(t *testing.T) {
	t.Run("return file errors early", func(t *testing.T) {
		testcases := []struct {
			label  string
			path   string
			expErr error
		}{
			{
				label:  "not found",
				path:   configPath("invalid/bad path"),
				expErr: configfile.ErrFileNotFound,
			},
			{
				label:  "unsupported extension",
				path:   configPath("invalid/badext.yams"),
				expErr: configfile.ErrFileExt,
			},
			{
				label:  "yaml invalid fields",
				path:   configPath("invalid/badfields.yml"),
				expErr: configfile.ErrFileParse,
			},
			{
				label:  "json invalid fields",
				path:   configPath("invalid/badfields.json"),
				expErr: configfile.ErrFileParse,
			},
			{
				label:  "self reference",
				path:   configPath("extends/extends-circular-self.yml"),
				expErr: configfile.ErrCircularExtends,
			},
			{
				label:  "circular reference",
				path:   configPath("extends/extends-circular-0.yml"),
				expErr: configfile.ErrCircularExtends,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				brunner := runner.Runner{}
				gotErr := configfile.Parse(tc.path, &brunner)

				if gotErr == nil {
					t.Fatal("exp non-nil error, got nil")
				}

				if !errors.Is(gotErr, tc.expErr) {
					t.Errorf("\nexp %v\ngot %v", tc.expErr, gotErr)
				}

				if !reflect.DeepEqual(brunner, runner.Runner{}) {
					t.Errorf("\nexp empty config\ngot %v", brunner)
				}
			})
		}
	})

	t.Run("happy path for all extensions", func(t *testing.T) {
		for _, ext := range supportedExt {
			expCfg := newExpConfig()
			fname := configPath("valid/benchttp" + ext)

			gotCfg := runner.Runner{}
			if err := configfile.Parse(fname, &gotCfg); err != nil {
				// critical error, stop the test
				t.Fatal(err)
			}

			if sameConfig(gotCfg, runner.Runner{}) {
				t.Error("received an empty configuration")
			}

			if !sameConfig(gotCfg, expCfg) {
				t.Errorf("unexpected parsed config for %s file:\nexp %#v\ngot %#v", ext, expCfg, gotCfg)
			}

		}
	})

	t.Run("override input config", func(t *testing.T) {
		brunner := runner.Runner{}
		brunner.Request = testutil.MustMakeRequest("POST", "https://overriden.com", nil, nil)
		brunner.GlobalTimeout = 10 * time.Millisecond

		fname := configPath("valid/benchttp-zeros.yml")

		if err := configfile.Parse(fname, &brunner); err != nil {
			t.Fatal(err)
		}

		const (
			expMethod        = "POST"                // from input config
			expGlobalTimeout = 42 * time.Millisecond // from read file
		)

		if gotMethod := brunner.Request.Method; gotMethod != expMethod {
			t.Errorf(
				"did not keep input values that are not set: "+
					"exp Request.Method == %s, got %s",
				expMethod, gotMethod,
			)
		}

		if gotGlobalTimeout := brunner.GlobalTimeout; gotGlobalTimeout != expGlobalTimeout {
			t.Errorf(
				"did not override input values that are set: "+
					"exp Runner.GlobalTimeout == %v, got %v",
				expGlobalTimeout, gotGlobalTimeout,
			)
		}

		t.Log(brunner)
	})

	t.Run("extend config files", func(t *testing.T) {
		testcases := []struct {
			label  string
			cfname string
			cfpath string
		}{
			{
				label:  "same directory",
				cfname: "child",
				cfpath: configPath("extends/extends-valid-child.yml"),
			},
			{
				label:  "nested directory",
				cfname: "nested",
				cfpath: configPath("extends/nest-0/nest-1/extends-valid-nested.yml"),
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				var brunner runner.Runner
				if err := configfile.Parse(tc.cfpath, &brunner); err != nil {
					t.Fatal(err)
				}

				var (
					expMethod = "POST"
					expURL    = fmt.Sprintf("http://%s.config", tc.cfname)
				)

				if gotMethod := brunner.Request.Method; gotMethod != expMethod {
					t.Errorf("method: exp %s, got %s", expMethod, gotMethod)
				}

				if gotURL := brunner.Request.URL.String(); gotURL != expURL {
					t.Errorf("url: exp %s, got %s", expURL, gotURL)
				}
			})
		}
	})
}

// helpers

// newExpConfig returns the expected runner.ConfigConfig result after parsing
// one of the config files in testdataConfigPath.
func newExpConfig() runner.Runner {
	return runner.Runner{
		Request: testutil.MustMakeRequest(
			"POST",
			validURL,
			http.Header{
				"key0": []string{"val0", "val1"},
				"key1": []string{"val0"},
			},
			[]byte(`{"key0":"val0","key1":"val1"}`),
		),

		Requests:       100,
		Concurrency:    1,
		Interval:       50 * time.Millisecond,
		RequestTimeout: 2 * time.Second,
		GlobalTimeout:  60 * time.Second,

		Tests: []runner.TestCase{
			{
				Name:      "minimum response time",
				Field:     "ResponseTimes.Min",
				Predicate: "GT",
				Target:    80 * time.Millisecond,
			},
			{
				Name:      "maximum response time",
				Field:     "ResponseTimes.Max",
				Predicate: "LTE",
				Target:    120 * time.Millisecond,
			},
			{
				Name:      "100% availability",
				Field:     "RequestFailureCount",
				Predicate: "EQ",
				Target:    0,
			},
		},
	}
}

func sameConfig(a, b runner.Runner) bool {
	if a.Request == nil || b.Request == nil {
		return a.Request == nil && b.Request == nil
	}
	return sameURL(a.Request.URL, b.Request.URL) &&
		sameHeader(a.Request.Header, b.Request.Header) &&
		sameBody(a.Request.Body, b.Request.Body)
}

// sameURL returns true if a and b are the same *url.URL, taking into account
// the undeterministic nature of their RawQuery.
func sameURL(a, b *url.URL) bool {
	// check query params equality via Query() rather than RawQuery
	if !reflect.DeepEqual(a.Query(), b.Query()) {
		return false
	}

	// temporarily set RawQuery to a determined value
	for _, u := range []*url.URL{a, b} {
		defer setTempValue(&u.RawQuery, "replaced by test")()
	}

	// we can now rely on deep equality check
	return reflect.DeepEqual(a, b)
}

func sameHeader(a, b http.Header) bool {
	return reflect.DeepEqual(a, b)
	// if len(a) != len(b) {
	// 	return false
	// }
	// for k, values := range a {
	// 	if len(values) != len()
	// }
}

func sameBody(a, b io.ReadCloser) bool {
	return reflect.DeepEqual(a, b)
}

// setTempValue sets *ptr to val and returns a restore func that sets *ptr
// back to its previous value.
func setTempValue(ptr *string, val string) (restore func()) {
	previousValue := *ptr
	*ptr = val
	return func() {
		*ptr = previousValue
	}
}

func configPath(name string) string {
	return filepath.Join(validConfigPath, name)
}
