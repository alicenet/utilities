package logz

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alicenet/indexer/internal/logz/severity"
)

// Set this sufficiently high or weird things happen when the buffer needs to be increased.
const defaultSize = 1 << 16

func setupEntry() (*bytes.Buffer, *entry) {
	buf := bytes.NewBuffer(make([]byte, defaultSize))
	l := &logger{
		encoder: json.NewEncoder(buf),
	}

	return buf, l.entry()
}

func TestPackageLogger(t *testing.T) { //nolint: paralleltest // Package level logger, can't be run in parallel.
	buf := bytes.NewBuffer(make([]byte, defaultSize))
	packageLogger.encoder = json.NewEncoder(buf)

	tests := []struct {
		fn  func(v ...any)
		fnf func(format string, v ...any)
		sev severity.Severity
	}{
		{Debug, Debugf, severity.Debug},
		{Info, Infof, severity.Info},
		{Notice, Noticef, severity.Notice},
		{Warning, Warningf, severity.Warning},
		{Error, Errorf, severity.Error},
		{Critical, Criticalf, severity.Critical},
		{Alert, Alertf, severity.Alert},
		{Emergency, Emergencyf, severity.Emergency},
	}

	for _, v := range tests { //nolint: paralleltest // Package level logger, can't be run in parallel.
		t.Run(string(v.sev), func(t *testing.T) {
			v.fn("hello", "world")
			result := buf.String()
			if !strings.Contains(result, "helloworld") {
				t.Errorf("log missing message: %s\n", result)
			}
			if !strings.Contains(result, string(v.sev)) {
				t.Errorf("log missing severity: %s\n", result)
			}
			buf.Reset()

			v.fnf("hello%s", "world")
			result = buf.String()
			if !strings.Contains(result, "helloworld") {
				t.Errorf("log missing message: %s\n", result)
			}
			if !strings.Contains(result, string(v.sev)) {
				t.Errorf("log missing severity: %s\n", result)
			}
			buf.Reset()
		})
	}
}

func TestBasicEntries(t *testing.T) { //nolint: tparallel // Shared buffer, can't be run in parallel.
	t.Parallel()

	buf, e := setupEntry()

	tests := []struct {
		fn  func(v ...any)
		fnf func(format string, v ...any)
		sev severity.Severity
	}{
		{e.Debug, e.Debugf, severity.Debug},
		{e.Info, e.Infof, severity.Info},
		{e.Notice, e.Noticef, severity.Notice},
		{e.Warning, e.Warningf, severity.Warning},
		{e.Error, e.Errorf, severity.Error},
		{e.Critical, e.Criticalf, severity.Critical},
		{e.Alert, e.Alertf, severity.Alert},
		{e.Emergency, e.Emergencyf, severity.Emergency},
	}

	for _, v := range tests { //nolint: paralleltest // Shared buffer, can't be run in parallel.
		t.Run(string(v.sev), func(t *testing.T) {
			v.fn("hello", "world")
			result := buf.String()
			if !strings.Contains(result, "helloworld") {
				t.Errorf("log missing message: %s\n", result)
			}
			if !strings.Contains(result, string(v.sev)) {
				t.Errorf("log missing severity: %s\n", result)
			}
			buf.Reset()

			v.fnf("hello%s", "world")
			result = buf.String()
			if !strings.Contains(result, "helloworld") {
				t.Errorf("log missing message: %s\n", result)
			}
			if !strings.Contains(result, string(v.sev)) {
				t.Errorf("log missing severity: %s\n", result)
			}
			buf.Reset()
		})
	}
}

func TestDetails(t *testing.T) {
	t.Parallel()

	buf, e := setupEntry()

	basic := e.WithDetail("aaa", "bbb")

	second := basic.WithDetails(Details{"ccc": 123, "ddd": 1.23})

	second.Info("testing")

	result := buf.String()

	if !strings.Contains(result, `"ccc":123`) {
		t.Errorf("log missing details: %s\n", result)
	}

	if !strings.Contains(result, `"ddd":1.23`) {
		t.Errorf("log missing details: %s\n", result)
	}

	buf.Reset()

	basic.Info("testing")

	result = buf.String()

	if !strings.Contains(result, `"aaa":"bbb"`) {
		t.Errorf("log missing details: %s\n", result)
	}

	if strings.Contains(result, `"ccc":123`) {
		t.Errorf("log contained unexpected detail: %s\n", result)
	}
}

func TestSourceLocation(t *testing.T) {
	t.Parallel()

	buf, e := setupEntry()

	// Running twice on same line to exercise source location caching logic.
	for i := 0; i < 2; i++ {
		e.Info("testing")
	}

	result := buf.String()

	if !strings.Contains(result, "logz.TestSourceLocation") {
		t.Error("log missing function")
	}

	if !strings.Contains(result, "logz_test.go") {
		t.Error("log missing filename")
	}
}
