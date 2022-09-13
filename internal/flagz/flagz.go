package flagz

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alicenet/indexer/internal/logz"
)

const (
	// exitCode specified by the flag package for invalid parsing.
	exitCode = 2
)

// Parse parses the command-line flags from os.Args[1:]. Must be called after all flags are defined
// and before flags are accessed by the program. It will override flags based on environment flags
// based on the uppercase name of the flag.
func Parse() {
	_ = ParseFlagSet(flag.CommandLine, os.Args[1:])
}

// ParseFlagSet definitions from the argument list, which should not include the command name.
// Must be called after all flags in the FlagSet are defined and before flags are accessed by the
// program. The return value will be ErrHelp if -help or -h were set but not defined.
func ParseFlagSet(flagset *flag.FlagSet, args []string) error {
	if err := parseFlagSet(flagset, args); err != nil {
		switch flagset.ErrorHandling() {
		case flag.ContinueOnError:
			return err
		case flag.ExitOnError:
			fmt.Fprintln(flagset.Output(), err)
			os.Exit(exitCode)
		case flag.PanicOnError:
			panic(err)
		}
	}

	return nil
}

// parseFlagSet helper function runs the actual environment overrides.
func parseFlagSet(flagset *flag.FlagSet, args []string) error {
	var err error

	flagset.VisitAll(func(f *flag.Flag) {
		// Fail fast on first error.
		if err != nil {
			return
		}
		err = apply(flagset, f.Name)

		logz.WithDetails(logz.Details{
			"flag": f,
		}).Debug("processing flag")
	})

	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	return flagset.Parse(args)
}

// apply environment variables to any matching flags.
func apply(fs *flag.FlagSet, name string) error {
	envName := strings.ToUpper(name)
	if v, ok := os.LookupEnv(envName); ok {
		if err := fs.Set(name, v); err != nil {
			return fmt.Errorf("set %s to %s: %w", name, v, err)
		}

		logz.WithDetails(logz.Details{"name": name, "value": v}).Debug("got environment override for flag")
	}

	return nil
}
