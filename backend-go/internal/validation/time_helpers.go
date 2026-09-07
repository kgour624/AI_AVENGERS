package validation

import "time"

// nowMs returns current time as Unix milliseconds.
func nowMs() int64 {
	return time.Now().UnixMilli()
}

// elapsedMs returns milliseconds elapsed since startMs.
func elapsedMs(startMs int64) int64 {
	return time.Now().UnixMilli() - startMs
}
