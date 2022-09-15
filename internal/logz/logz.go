package logz

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/alicenet/utilities/internal/logz/severity"
)

// A Logger that outputs structured logs.
type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Notice(v ...any)
	Noticef(format string, v ...any)
	Warning(v ...any)
	Warningf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Critical(v ...any)
	Criticalf(format string, v ...any)
	Alert(v ...any)
	Alertf(format string, v ...any)
	Emergency(v ...any)
	Emergencyf(format string, v ...any)
	WithDetail(key string, value any) Logger
	WithDetails(Details) Logger
}

type Details map[string]any

// sourceLocation of where the entry was logged from in code.
type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Function string `json:"function,omitempty"`
	Line     string `json:"line,omitempty"`
}

// An entry to be logged, carries all necessary details.
type entry struct {
	logger   *logger
	Message  string            `json:"message,omitempty"`
	Severity severity.Severity `json:"severity,omitempty"`
	//nolint: tagliatelle // JSON name from Stackdriver standard.
	SourceLocation *sourceLocation `json:"logging.googleapis.com/sourceLocation,omitempty"`
	Details        Details         `json:"details,omitempty"`
	// Timestamp is intentionally excluded as it will get marked with system time automatically.
}

// clone an entry so that modifications to the clone don't impact the original.
func (e *entry) clone() *entry {
	return &entry{
		logger:  e.logger,
		Details: maps.Clone(e.Details),
	}
}

// log a message with severity and details associated.
func (e *entry) log(msg string, sev severity.Severity) {
	e.logger.Lock()
	defer e.logger.Unlock()
	e.Message = msg
	e.Severity = sev

	e.SourceLocation = sources.get()

	if err := e.logger.encoder.Encode(e); err != nil {
		panic(err)
	}
}

var (
	//nolint: gochecknoglobals // Needed for package level functionality.
	packageLogger = &logger{
		encoder: json.NewEncoder(os.Stdout),
	}
	//nolint: gochecknoglobals // Needed for package level functionality.
	sources = &sourcesMap{
		sources: make(map[uintptr]*sourceLocation),
	}
	//nolint: gochecknoglobals // Needed for package level functionality.
	packageEntry = packageLogger.entry()
)

// sourcesMap to amortize the lookup cost for source locations.
type sourcesMap struct {
	sync.Mutex
	sources map[uintptr]*sourceLocation
}

// get the sourceLocation where a log was called from. Will cache.
func (s *sourcesMap) get() *sourceLocation {
	const depth = 3

	s.Lock()
	defer s.Unlock()

	pc, file, line, ok := runtime.Caller(depth)

	if !ok {
		panic("could not get source location")
	}

	if source, ok := s.sources[pc]; ok {
		return source
	}

	source := &sourceLocation{
		File: file,
		Line: strconv.Itoa(line),
	}

	fn := runtime.FuncForPC(pc)
	if fn != nil {
		source.Function = fn.Name()
	}

	s.sources[pc] = source

	return source
}

// a logger is a threadsafe means of outputting structured logs.
type logger struct {
	sync.Mutex
	encoder *json.Encoder
}

// entry created from a logger.
func (l *logger) entry() *entry {
	e := &entry{
		logger:  l,
		Details: make(Details),
	}

	return e
}

func Debug(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Debug)
}

func (e *entry) Debug(v ...any) {
	e.log(fmt.Sprint(v...), severity.Debug)
}

func Debugf(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Debug)
}

func (e *entry) Debugf(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Debug)
}

func Info(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Info)
}

func (e *entry) Info(v ...any) {
	e.log(fmt.Sprint(v...), severity.Info)
}

func Infof(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Info)
}

func (e *entry) Infof(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Info)
}

func Notice(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Notice)
}

func (e *entry) Notice(v ...any) {
	e.log(fmt.Sprint(v...), severity.Notice)
}

func Noticef(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Notice)
}

func (e *entry) Noticef(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Notice)
}

func Warning(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Warning)
}

func (e *entry) Warning(v ...any) {
	e.log(fmt.Sprint(v...), severity.Warning)
}

func Warningf(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Warning)
}

func (e *entry) Warningf(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Warning)
}

func Error(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Error)
}

func (e *entry) Error(v ...any) {
	e.log(fmt.Sprint(v...), severity.Error)
}

func Errorf(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Error)
}

func (e *entry) Errorf(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Error)
}

func Critical(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Critical)
}

func (e *entry) Critical(v ...any) {
	e.log(fmt.Sprint(v...), severity.Critical)
}

func Criticalf(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Critical)
}

func (e *entry) Criticalf(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Critical)
}

func Alert(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Alert)
}

func (e *entry) Alert(v ...any) {
	e.log(fmt.Sprint(v...), severity.Alert)
}

func Alertf(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Alert)
}

func (e *entry) Alertf(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Alert)
}

func Emergency(v ...any) {
	packageEntry.log(fmt.Sprint(v...), severity.Emergency)
}

func (e *entry) Emergency(v ...any) {
	e.log(fmt.Sprint(v...), severity.Emergency)
}

func Emergencyf(format string, v ...any) {
	packageEntry.log(fmt.Sprintf(format, v...), severity.Emergency)
}

func (e *entry) Emergencyf(format string, v ...any) {
	e.log(fmt.Sprintf(format, v...), severity.Emergency)
}

func WithDetail(key string, value any) Logger {
	entry := packageEntry.clone()
	entry.Details[key] = value

	return entry
}

func (e *entry) WithDetail(key string, value any) Logger {
	entry := e.clone()
	entry.Details[key] = value

	return entry
}

func WithDetails(details Details) Logger {
	entry := packageEntry.clone()

	for k, v := range details {
		entry.Details[k] = v
	}

	return entry
}

func (e *entry) WithDetails(details Details) Logger {
	entry := e.clone()

	for k, v := range details {
		entry.Details[k] = v
	}

	return entry
}
