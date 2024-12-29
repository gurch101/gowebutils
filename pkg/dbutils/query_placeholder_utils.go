package dbutils

import (
	"fmt"
	"strings"
)

// generatePlaceholders creates a SQL-compatible placeholder string for a given number of fields.
func generatePlaceholders(rowNum, numFields int) string {
	placeholders := make([]string, numFields)
	for i := 0; i < numFields; i++ {
		placeholders[i] = fmt.Sprintf("$%d", (rowNum*numFields)+i+1)
	}
	return fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
}

// GetChunkSize calculates the number of chunks needed to process a given number of rows and fields.
// The chunks must fit within the database limit of 65,000 placeholders.
func GetChunkSize(numRows, numFields int) int {
	// Maximum number of placeholders allowed by the database
	maxPlaceholders := 65000

	// Calculate the number of rows that can fit in a single chunk
	rowsPerChunk := maxPlaceholders / numFields

	return min(numRows, rowsPerChunk)
}

// ProcessInChunks splits a large set of arguments into manageable chunks and applies a callback to each chunk.
// This ensures SQL statements do not exceed database limits.
//
// Args:
//
//	args: The full list of arguments to process.
//	numFields: The number of fields in each placeholder tuple.
//	callback: A function to handle each chunk of arguments. It takes the chunk and its placeholder string as input.
//
// Returns:
//
//	An error if the callback fails for any chunk.
func ProcessInChunks[T any](args []T, chunkSize, numFields int, callback func(chunk []T, placeholders string) error) error {
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}
	if numFields <= 0 {
		return fmt.Errorf("invalid number of fields: %d", numFields)
	}

	if chunkSize == len(args) {
		placeholders := make([]string, 0, len(args))
		for i := 0; i < len(args); i++ {
			placeholders = append(placeholders, generatePlaceholders(i, numFields))
		}
		return callback(args, strings.Join(placeholders, ","))
	}

	placeholders := make([]string, 0, chunkSize)
	chunk := make([]T, 0, chunkSize)

	for i := 0; i < len(args); i++ {
		if len(chunk) >= chunkSize {
			// Process the current chunk with the callback
			if err := callback(chunk, strings.Join(placeholders, ",")); err != nil {
				return err
			}
			// Reset chunk and placeholders for the next batch
			chunk = make([]T, 0, chunkSize)
			placeholders = make([]string, 0, chunkSize)
		}
		// Add the current argument and placeholder to the current batch
		chunk = append(chunk, args[i])
		placeholders = append(placeholders, generatePlaceholders(i, numFields))
	}

	// Process the remaining items in the final chunk
	return callback(chunk, strings.Join(placeholders, ","))
}
