package configfile

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/benchttp/engine/configparse"
	"github.com/benchttp/engine/runner"

	"github.com/benchttp/cli/internal/errorutil"
)

// Parse parses a benchttp runner config file into a runner.ConfigGlobal
// and returns it or the first non-nil error occurring in the process,
// which can be any of the values declared in the package.
func Parse(filename string) (cfg runner.Config, err error) {
	uconfs, err := parseFileRecursive(filename, []configparse.Representation{}, set{})
	if err != nil {
		return
	}
	return parseAndMergeConfigs(uconfs)
}

// set is a collection of unique string values.
type set map[string]bool

// add adds v to the receiver. If v is already set, it returns a non-nil
// error instead.
func (s set) add(v string) error {
	if _, exists := s[v]; exists {
		return errors.New("value already set")
	}
	s[v] = true
	return nil
}

// parseFileRecursive parses a config file and its parent found from key
// "extends" recursively until the root config file is reached.
// It returns the list of all parsed configs or the first non-nil error
// occurring in the process.
func parseFileRecursive(
	filename string,
	reprs []configparse.Representation,
	seen set,
) ([]configparse.Representation, error) {
	// avoid infinite recursion caused by circular reference
	if err := seen.add(filename); err != nil {
		return reprs, ErrCircularExtends
	}

	// parse current file, append parsed config
	repr, err := parseFile(filename)
	if err != nil {
		return reprs, err
	}
	reprs = append(reprs, repr)

	// root config reached: stop now and return the parsed configs
	if repr.Extends == nil {
		return reprs, nil
	}

	// config has parent: resolve its path and parse it recursively
	parentPath := filepath.Join(filepath.Dir(filename), *repr.Extends)
	return parseFileRecursive(parentPath, reprs, seen)
}

// parseFile parses a single config file and returns the result as an
// configparse.Representation and an appropriate error predeclared in the package.
func parseFile(filename string) (repr configparse.Representation, err error) {
	b, err := os.ReadFile(filename)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return repr, errorutil.WithDetails(ErrFileNotFound, filename)
	default:
		return repr, errorutil.WithDetails(ErrFileRead, filename, err)
	}

	ext := extension(filepath.Ext(filename))
	parser, err := newParser(ext)
	if err != nil {
		return repr, errorutil.WithDetails(ErrFileExt, ext, err)
	}

	if err = parser.Parse(b, &repr); err != nil {
		return repr, errorutil.WithDetails(ErrParse, filename, err)
	}

	return repr, nil
}

// parseAndMergeConfigs iterates backwards over uconfs, parsing them
// as runner.ConfigGlobal and merging them into a single one.
// It returns the merged result or the first non-nil error occurring in the
// process.
func parseAndMergeConfigs(reprs []configparse.Representation) (cfg runner.Config, err error) {
	if len(reprs) == 0 { // supposedly catched upstream, should not occur
		return cfg, errors.New(
			"an unacceptable error occurred parsing the config file, " +
				"please visit https://github.com/benchttp/runner/issues/new " +
				"and insult us properly",
		)
	}

	cfg = runner.DefaultConfig()

	for i := len(reprs) - 1; i >= 0; i-- {
		repr := reprs[i]
		currentConfig, err := configparse.ParseRepresentation(repr)
		if err != nil {
			return cfg, errorutil.WithDetails(ErrParse, err)
		}
		cfg = currentConfig.Override(cfg)
	}

	return cfg, nil
}
