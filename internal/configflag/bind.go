package configflag

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/benchttp/engine/configparse"
)

// Bind reads arguments provided to flagset as config fields
// and binds their value to the appropriate fields of dst.
// The provided *flag.Flagset must not have been parsed yet, otherwise
// bindings its values would fail.
func Bind(flagset *flag.FlagSet, dst *configparse.Representation) {
	for field, bind := range bindings {
		flagset.Func(field, flagsUsage[field], bind(dst))
	}
}

type (
	repr       = configparse.Representation
	flagSetter = func(string) error
)

var bindings = map[string]func(*repr) flagSetter{
	flagMethod: func(dst *repr) flagSetter {
		return func(in string) error {
			dst.Request.Method = &in
			return nil
		}
	},
	flagURL: func(dst *repr) flagSetter {
		return func(in string) error {
			dst.Request.URL = &in
			return nil
		}
	},
	flagHeader: func(dst *repr) flagSetter {
		return func(in string) error {
			keyval := strings.SplitN(in, ":", 2)
			if len(keyval) != 2 {
				return errors.New(`-header: expect format "<key>:<value>"`)
			}
			key, val := keyval[0], keyval[1]
			if dst.Request.Header == nil {
				dst.Request.Header = map[string][]string{}
			}
			dst.Request.Header[key] = append(dst.Request.Header[key], val)
			return nil
		}
	},
	flagBody: func(dst *repr) flagSetter {
		return func(in string) error {
			errFormat := fmt.Errorf(`expect format "<type>:<content>", got %q`, in)
			if in == "" {
				return errFormat
			}
			split := strings.SplitN(in, ":", 2)
			if len(split) != 2 {
				return errFormat
			}
			btype, bcontent := split[0], split[1]
			if bcontent == "" {
				return errFormat
			}
			switch btype {
			case "raw":
				dst.Request.Body = &struct {
					Type    string `yaml:"type" json:"type"`
					Content string `yaml:"content" json:"content"`
				}{
					Type:    btype,
					Content: bcontent,
				}
			// case "file":
			// 	// TODO
			default:
				return fmt.Errorf(`unsupported type: %s (only "raw" accepted)`, btype)
			}
			return nil
		}
	},
	flagRequests: func(dst *repr) flagSetter {
		return func(in string) error {
			n, err := strconv.Atoi(in)
			if err != nil {
				return err
			}
			dst.Runner.Requests = &n
			return nil
		}
	},
	flagConcurrency: func(dst *repr) flagSetter {
		return func(in string) error {
			n, err := strconv.Atoi(in)
			if err != nil {
				return err
			}
			dst.Runner.Concurrency = &n
			return nil
		}
	},
	flagInterval: func(dst *repr) flagSetter {
		return func(in string) error {
			dst.Runner.Interval = &in
			return nil
		}
	},
	flagRequestTimeout: func(dst *repr) flagSetter {
		return func(in string) error {
			dst.Runner.RequestTimeout = &in
			return nil
		}
	},
	flagGlobalTimeout: func(dst *repr) flagSetter {
		return func(in string) error {
			dst.Runner.GlobalTimeout = &in
			return nil
		}
	},
}
