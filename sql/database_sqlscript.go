package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// DataBaseExecuteScript executes a SQL script in the specified database
func (c *Connector) DataBaseExecuteScript(ctx context.Context, database string, script string) error {
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(script)), "SELECT") {
		var exists int
		err := c.
			setDatabase(&database).
			QueryRowContext(ctx, script,
				func(r *sql.Row) error {
					return r.Scan(&exists)
				},
			)
		if err == sql.ErrNoRows {
			return fmt.Errorf("no rows returned from verification query")
		}
		return err
	}

	// Split the script into batches
	batches := splitBatches(script)
	
	// Execute each batch
	for _, batch := range batches {
		// Prepare the dynamic SQL execution command for this batch
		cmd := `DECLARE @stmt nvarchar(max)
				SET @stmt = @script
				EXEC sp_executesql @stmt`

		// Execute the batch
		err := c.
			setDatabase(&database).
			ExecContext(ctx, cmd,
				sql.Named("script", batch),
				sql.Named("database", database),
			)
		if err != nil {
			return errors.Wrapf(err, "failed to execute batch: %s", truncateString(batch, 100))
		}
	}

	return nil
}

// splitBatches splits a SQL script into individual batches based on GO statements
func splitBatches(script string) []string {
	// First normalize line endings
	script = strings.ReplaceAll(script, "\r\n", "\n")
	
	var batches []string
	lines := strings.Split(script, "\n")
	currentBatch := []string{}
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Check if the line is a GO statement (case insensitive, allowing for GO n syntax)
		if goMatch := strings.HasPrefix(strings.ToUpper(trimmedLine), "GO"); goMatch {
			// If we have statements in the current batch, add them
			if batchContent := strings.TrimSpace(strings.Join(currentBatch, "\n")); batchContent != "" {
				batches = append(batches, batchContent)
			}
			// Reset the current batch
			currentBatch = []string{}
			continue
		}
		
		currentBatch = append(currentBatch, line)
	}
	
	// Add the last batch if it's not empty
	if batchContent := strings.TrimSpace(strings.Join(currentBatch, "\n")); batchContent != "" {
		batches = append(batches, batchContent)
	}
	
	return batches
}

// truncateString truncates a string to the specified length and adds "..." if truncated
func truncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}
