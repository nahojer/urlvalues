package urlvalues

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseTime(layout, value string) (time.Time, error) {
	now := time.Now()
	if value == "now" {
		return now, nil
	}

	// Parse time based on now. For example, "now-2y+3m-2d" subtracts 2 years,
	// adds 3 months and subtracts 3 days from now.
	if strings.HasPrefix(value, "now") {
		var years, months, days int

		// Remove "now" prefix from value.
		value = strings.TrimSpace(value[3:])

		// Separate parts by space so that we can split on it.
		value = strings.TrimSpace(strings.NewReplacer("+", " +", "-", " -").Replace(value))

		parts := strings.Split(value, " ")
		for _, part := range parts {
			// A part consist of at least 3 characters: a sign (+-), followed by one or
			// more digits, followed by a year/month/day (y/m/d) identifier.
			if len(part) < 3 {
				return time.Time{}, errors.New("invalid \"now\" based format")
			}

			var sign int
			switch part[0] {
			case '+':
				sign = 1
			case '-':
				sign = -1
			default:
				return time.Time{}, fmt.Errorf("invalid sign %q", part[0])
			}

			v, err := strconv.Atoi(part[1 : len(part)-1])
			if err != nil {
				return time.Time{}, fmt.Errorf("%q is not a valid integer", part[1:len(part)-1])
			}

			switch part[len(part)-1] {
			case 'y':
				years = sign * v
			case 'm':
				months = sign * v
			case 'd':
				days = sign * v
			default:
				return time.Time{}, fmt.Errorf("invalid year/month/day identifier %q", part[len(part)-1])
			}
		}

		return now.AddDate(years, months, days), nil
	}

	// Allow custom layouts. Valid layouts include the predefined layout constants in the
	// time package, as well as custom layouts defined by the consumer that time.Parse
	// understands. Defaults to time.Layout.
	switch layout {
	case "", "Layout":
		return time.Parse(time.Layout, value)
	case "ANSIC":
		return time.Parse(time.ANSIC, value)
	case "UnixDate":
		return time.Parse(time.UnixDate, value)
	case "RubyDate":
		return time.Parse(time.RubyDate, value)
	case "RFC822":
		return time.Parse(time.RFC822, value)
	case "RFC822Z":
		return time.Parse(time.RFC822Z, value)
	case "RFC850":
		return time.Parse(time.RFC850, value)
	case "RFC1123":
		return time.Parse(time.RFC1123, value)
	case "RFC1123Z":
		return time.Parse(time.RFC1123Z, value)
	case "RFC3339":
		return time.Parse(time.RFC3339, value)
	case "RFC3339Nano":
		return time.Parse(time.RFC3339Nano, value)
	case "Kitchen":
		return time.Parse(time.Kitchen, value)
	default:
		return time.Parse(layout, value)
	}
}
