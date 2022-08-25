package flagz

import (
	"flag"
	"strings"
	"testing"
)

const (
	testName  = "testing"
	flag1name = "testFlag1"
	flag2name = "testFlag2"
	unset     = "default"
	set       = "overridden"
)

func TestRegularFlags(t *testing.T) {
	t.Parallel()

	flagset := flag.NewFlagSet(testName, flag.PanicOnError)
	flag1 := flagset.String(flag1name, unset, testName)
	flag2 := flagset.String(flag2name, unset, testName)

	if err := ParseFlagSet(flagset, []string{"-testFlag2", set}); err != nil {
		t.Fatal(err)
	}

	if *flag1 != unset {
		t.Errorf("want: %s, got: %s", unset, *flag1)
	}

	if *flag2 != set {
		t.Errorf("want: %s, got: %s", set, *flag2)
	}
}

//nolint:paralleltest // t.Parallel not supported with t.Setenv
func TestEnvironmentOverride(t *testing.T) {
	flagset := flag.NewFlagSet(testName, flag.PanicOnError)
	flag1 := flagset.String(flag1name, unset, testName)

	t.Setenv(strings.ToUpper(flag1name), set)

	if err := ParseFlagSet(flagset, nil); err != nil {
		t.Fatal(err)
	}

	if *flag1 != set {
		t.Errorf("want: %s, got: %s", set, *flag1)
	}
}

func TestBadFlagParse(t *testing.T) {
	t.Parallel()

	flagset := flag.NewFlagSet(testName, flag.ContinueOnError)
	_ = flagset.Int64(flag1name, 0, testName)

	if err := ParseFlagSet(flagset, []string{"-testFlag1", set}); err == nil {
		t.Error("expected error but there was none")
	}
}

//nolint:paralleltest // t.Parallel not supported with t.Setenv
func TestBadEnvironmentParse(t *testing.T) {
	flagset := flag.NewFlagSet(testName, flag.ContinueOnError)

	_ = flagset.Int64(flag1name, 0, testName)
	_ = flagset.Int64(flag2name, 0, testName)

	t.Setenv(strings.ToUpper(flag1name), set)

	if err := ParseFlagSet(flagset, nil); err == nil {
		t.Error("expected error but there was none")
	}
}

//nolint:paralleltest // t.Parallel not supported with t.Setenv
func TestErrorHandlingRespected(t *testing.T) {
	// Intentially ignoring flag.ExitOnError as it calls os.Exit.
	flagset := flag.NewFlagSet(testName, flag.PanicOnError)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but there was none")
		}
	}()

	_ = flagset.Int64(flag1name, 0, testName)

	t.Setenv(strings.ToUpper(flag1name), set)

	_ = ParseFlagSet(flagset, nil)
}
