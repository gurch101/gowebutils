package fsutils

import (
	"fmt"
	"io"
)

func CloseAndPanic(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(fmt.Sprintf("failed to close: %v", err))
	}
}
