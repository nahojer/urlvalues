// Bridge package to expose urlvalues internals to tests in the urlvalues_test
// package.

package urlvalues

import "time"

func ExportParseTime(layout, value string) (time.Time, error) { return parseTime(layout, value) }
