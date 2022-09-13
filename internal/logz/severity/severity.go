package severity

type Severity string

// severity levels as defined in:
// - https://github.com/googleapis/googleapis/blob/master/google/logging/type/log_severity.proto
// - https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
const (
	Default   Severity = "DEFAULT"   // The log entry has no assigned severity level.
	Debug     Severity = "DEBUG"     // Debug or trace information.
	Info      Severity = "INFO"      // Routine information, such as ongoing status or performance.
	Notice    Severity = "NOTICE"    // Normal but significant events, such as start up, shut down, or a config change.
	Warning   Severity = "WARNING"   // Warning events might cause problems.
	Error     Severity = "ERROR"     // Error events are likely to cause problems.
	Critical  Severity = "CRITICAL"  // Critical events cause more severe problems or outages.
	Alert     Severity = "ALERT"     // A person must take an action immediately.
	Emergency Severity = "EMERGENCY" // One or more systems are unusable.
)
