package urlvalues_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nahojer/urlvalues"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name          string
		layout, value string
		want          time.Time
	}{
		// "now" based parsing.
		{"now", "", "now", time.Now()},
		{"now-1y", "", "now-1y", time.Now().AddDate(-1, 0, 0)},
		{"now+2y", "", "now+2y", time.Now().AddDate(2, 0, 0)},
		{"now-3m", "", "now-3m", time.Now().AddDate(0, -3, 0)},
		{"now+4m", "", "now+4m", time.Now().AddDate(0, 4, 0)},
		{"now-5d", "", "now-5d", time.Now().AddDate(0, 0, -5)},
		{"now+6d", "", "now+6d", time.Now().AddDate(0, 0, 6)},
		{"now+1y+1m+1d", "", "now+1y+1m+1d", time.Now().AddDate(1, 1, 1)},
		{"now-1y-1m-1d", "", "now-1y-1m-1d", time.Now().AddDate(-1, -1, -1)},
		{"now-9d-2y+5m", "", "now-9d-2y+5m", time.Now().AddDate(-2, 5, -9)},
		{"now+3y-2m", "", "now+3y-2m", time.Now().AddDate(3, -2, 0)},
		{"now+2m-7d", "", "now+2m-7d", time.Now().AddDate(0, 2, -7)},
		{"now+2y+5d", "", "now+2y+5d", time.Now().AddDate(2, 0, 5)},
		// Layout based parsing.
		{"default layout", "", time.Layout, parseTime(t, time.Layout, time.Layout)},
		{"custom layout", "2006-01-02", "2006-01-02", parseTime(t, "2006-01-02", "2006-01-02")},
		{"Layout", "Layout", time.Layout, parseTime(t, time.Layout, time.Layout)},
		{"ANSIC", "ANSIC", "Mon Jan 22 15:04:05 2006", parseTime(t, time.ANSIC, "Mon Jan 22 15:04:05 2006")},
		{"UnixDate", "UnixDate", "Mon Jan 22 15:04:05 MST 2006", parseTime(t, time.UnixDate, "Mon Jan 22 15:04:05 MST 2006")},
		{"RubyDate", "RubyDate", time.RubyDate, parseTime(t, time.RubyDate, time.RubyDate)},
		{"RFC822", "RFC822", time.RFC822, parseTime(t, time.RFC822, time.RFC822)},
		{"RFC822Z", "RFC822Z", time.RFC822Z, parseTime(t, time.RFC822Z, time.RFC822Z)},
		{"RFC850", "RFC850", time.RFC850, parseTime(t, time.RFC850, time.RFC850)},
		{"RFC1123", "RFC1123", time.RFC1123, parseTime(t, time.RFC1123, time.RFC1123)},
		{"RFC1123Z", "RFC1123Z", time.RFC1123Z, parseTime(t, time.RFC1123Z, time.RFC1123Z)},
		{"RFC3339 variant 1", "RFC3339", "2006-01-02T15:04:05Z", parseTime(t, time.RFC3339, "2006-01-02T15:04:05Z")},
		{"RFC3339 variant 2", "RFC3339", "2006-01-02T15:04:05+07:00", parseTime(t, time.RFC3339, "2006-01-02T15:04:05+07:00")},
		{"RFC3339Nano variant 1", "RFC3339Nano", "2006-01-02T15:04:05.999999999Z", parseTime(t, time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")},
		{"RFC3339Nano variant 2", "RFC3339Nano", "2006-01-02T15:04:05.999999999+07:00", parseTime(t, time.RFC3339Nano, "2006-01-02T15:04:05.999999999+07:00")},
		{"Kitchen", "Kitchen", time.Kitchen, parseTime(t, time.Kitchen, time.Kitchen)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := urlvalues.ExportParseTime(tt.layout, tt.value)
			if err != nil {
				t.Fatalf("urlvalues.parseTime(%q, %q) = %q, want <nil>", tt.layout, tt.value, err)
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Millisecond)) {
				t.Errorf("urlvalues.parseTime(%q, %q) -got +want\n%s", tt.layout, tt.value, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestParseTime_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		layout, value string
	}{
		{"empty", "", ""},
		{"invalid integer 1aa", "", "now-1aad"},
		{"nvalid sign <", "", "now<1y+1m+1d"},
		{"invalid identifier <", "", "now+1y+1m+1z"},
		{"wrong layout", "RFC822", "2006-01-02"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := urlvalues.ExportParseTime(tt.layout, tt.value); err == nil {
				t.Fatalf("urlvalues.parseTime(%q, %q) = %v, want error", tt.layout, tt.value, got)
			}
		})
	}
}

func parseTime(t *testing.T, layout string, value string) time.Time {
	t.Helper()

	tim, err := time.Parse(layout, value)
	if err != nil {
		t.Fatal(err)
	}

	return tim
}
