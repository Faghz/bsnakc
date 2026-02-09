package ports

import "context"

// Logger is the application-layer port for structured logging.
// Domain layer MUST NOT depend on this interface.
// Application services and adapters inject this via constructors.
//
// This interface is context-aware to support future OpenTelemetry integration.
// When OTel is configured, implementations can extract trace IDs from context
// and automatically include them in log entries.
//
// Usage:
//   - Application layer: log intent, decisions, business events
//   - Adapter layer: log technical details, external calls, errors
//   - Domain layer: NEVER log - communicate via errors or domain events only
type Logger interface {
	// Debug logs detailed diagnostic information (disabled in production).
	// Use for query execution details, detailed traces, debugging.
	Debug(ctx context.Context, msg string, fields ...Field)

	// Info logs significant business events or successful operations.
	// Use for user registrations, logins, state changes.
	Info(ctx context.Context, msg string, fields ...Field)

	// Warn logs recoverable errors or unexpected but handled situations.
	// Use for deprecated code paths, auto-corrections, graceful degradations.
	Warn(ctx context.Context, msg string, fields ...Field)

	// Error logs unexpected failures that indicate system problems.
	// Use for unhandled errors, integration failures, data inconsistencies.
	// Errors should be logged ONCE at the system boundary, not in lower layers.
	Error(ctx context.Context, msg string, fields ...Field)

	// With creates a child logger with persistent fields attached.
	// Useful for adding context like user_id, request_id to all subsequent logs.
	//
	// Example:
	//   userLogger := logger.With(String("user_id", id), String("domain", "users"))
	//   userLogger.Info(ctx, "User updated")  // Automatically includes user_id
	With(fields ...Field) Logger
}

// Field represents a structured log field (key-value pair).
// Implementations map this to their underlying logging library's field type.
type Field struct {
	Key   string
	Value interface{}
	Type  FieldType
}

// FieldType indicates how the field value should be serialized.
type FieldType uint8

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Float64Type
	BoolType
	DurationType
	TimeType
	ErrorType
	StackTraceType
	AnyType
)

// Field constructor helpers - use these to create structured log fields.

// String constructs a string field.
func String(key, val string) Field {
	return Field{Key: key, Value: val, Type: StringType}
}

// Int constructs an int field.
func Int(key string, val int) Field {
	return Field{Key: key, Value: val, Type: IntType}
}

// Int64 constructs an int64 field.
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val, Type: Int64Type}
}

// Float64 constructs a float64 field.
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val, Type: Float64Type}
}

// Bool constructs a boolean field.
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val, Type: BoolType}
}

// Duration constructs a time.Duration field (serialized as nanoseconds or human-readable).
func Duration(key string, val interface{}) Field {
	return Field{Key: key, Value: val, Type: DurationType}
}

// Time constructs a time.Time field.
func Time(key string, val interface{}) Field {
	return Field{Key: key, Value: val, Type: TimeType}
}

// Error constructs an error field (preserves stack traces if available).
func Error(err error) Field {
	return Field{Key: "error", Value: err, Type: ErrorType}
}

// StackTrace constructs a field that captures the current stack trace.
// Use this explicitly when you need stack trace information in logs.
// Example: logger.Error(ctx, "Critical failure", StackTrace(), Error(err))
func StackTrace() Field {
	return Field{Key: "stacktrace", Value: nil, Type: StackTraceType}
}

// Any constructs a field with arbitrary value (use sparingly - prefer typed helpers).
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val, Type: AnyType}
}
