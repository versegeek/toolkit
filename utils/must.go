package utils

// Must unwraps (value, error) pairs from functions that follow the standard
// Go error return convention. It panics with the error if err != nil.
// Intended for use in main() initialization or test helpers where a failure is
// unrecoverable.
//
// Example:
//
//	db := utils.Must(sql.Open("sqlite3", ":memory:"))
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Must0 panics if err is non-nil. Use for functions that return only an error.
//
// Example:
//
//	utils.Must0(db.Ping())
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}
