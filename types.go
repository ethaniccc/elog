package elog

const (
	LoggerModeTypeInfo uint64 = 1 << iota
	LoggerModeTypeWarning
	LoggerModeTypeError
	LoggerModeTypeFatal
	LoggerModeTypeDebug
)

const (
	LoggerOptIncludeTime uint64 = 1 << iota
)

var Modes = map[uint64]string{
	LoggerModeTypeInfo:    "info",
	LoggerModeTypeWarning: "warning",
	LoggerModeTypeError:   "error",
	LoggerModeTypeFatal:   "fatal",
	LoggerModeTypeDebug:   "debug",
}

// RegisterMode registers a logger type into the types map.
func RegisterMode(t uint64, v string) bool {
	if _, ok := Modes[t]; ok {
		return false
	}

	Modes[t] = v
	return true
}

// WithFlags returns a final type flag with all the types that the logger should use.
func WithFlags(types ...uint64) uint64 {
	var final_type uint64
	for _, t := range types {
		final_type = final_type | t
	}

	return final_type
}

func HasType(types, t uint64) bool {
	return types&t > 0
}
