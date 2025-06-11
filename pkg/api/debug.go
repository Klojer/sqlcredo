package api

type (
	// DebugFunc is a function type used for SQL query debugging and logging.
	// It receives the SQL query string and its arguments before execution.
	//
	// Parameters:
	//   - sql: The SQL query string to be executed
	//   - args: Variable number of query arguments of any type
	DebugFunc func(sql string, args ...any)
)
