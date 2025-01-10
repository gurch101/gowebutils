package threads

import (
	"fmt"
	"log/slog"
)

// Launch a background goroutine.
func Background(callback func()) {
	go func() {
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				slog.Error("background process failed", "error", fmt.Sprintf("recover panic: %v", err))
			}
		}()

		// Execute the arbitrary function that we passed as the parameter.
		callback()
	}()
}
