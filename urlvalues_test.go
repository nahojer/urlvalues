package urlvalues_test

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nahojer/urlvalues"
)

// TestTarget is the struct we use to test correctness of urlvalues.Unmarshal
// for all supported types.
type TestTarget struct {
	String     string `urlvalue:"aString,default:banana"`
	Ptr        *string
	PtrDefault *string `urlvalue:",default:orange"`
	Uint       uint
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32
	Uint64     uint64 `urlvalue:"aUint64,default:7"`
	Int        int    `urlvalue:"aInt,default:42"`
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64 `urlvalue:"aInt64,default:66"`
	Float32    float32
	Float64    float64           `urlvalue:"aFloat64,default:3.14159"`
	Bool       bool              `urlvalue:"aBool,default:true"`
	Skip       string            `urlvalue:"-"`
	NotSkipped string            `urlvalue:"-,"`
	Time       time.Time         `urlvalue:",default:now-1y+2m-10d"`
	Duration   time.Duration     `urlvalue:",default:1s"`
	Map        map[string]string `urlvalue:"aMap"`
	Anon       struct {
		Slice []string `urlvalue:"items"`
	}
	Embed
	TextUnmarshaler   TestTextUnmarshaler   `urlvalue:",default:text"`
	BinaryUnmarshaler TestBinaryUnmarshaler `urlvalue:",default:binary"`
}

type Embed struct {
	Byte byte
}

// TestTextUnmarshaler implements encoding.TextUnmarshaler.
type TestTextUnmarshaler string

func (t *TestTextUnmarshaler) UnmarshalText(text []byte) error {
	*t = TestTextUnmarshaler(fmt.Sprintf("text_%s", text))
	return nil
}

// TestBinaryUnmarshaler implements encoding.BinaryUnmarshaler.
type TestBinaryUnmarshaler []byte

func (b *TestBinaryUnmarshaler) UnmarshalBinary(data []byte) error {
	*b = TestBinaryUnmarshaler([]byte(fmt.Sprintf("data_%s", data)))
	return nil
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		in   url.Values
		want TestTarget
	}{
		{
			"nil data",
			nil,
			TestTarget{
				String: "banana", Ptr: nil, PtrDefault: ptr("orange"),
				Uint64: 7, Int: 42, Int64: 66, Float64: 3.14159,
				Bool: true, Time: time.Now().AddDate(-1, 2, -10), Duration: time.Second,
				TextUnmarshaler: "text_text", BinaryUnmarshaler: []byte("data_binary"),
			},
		},
		{
			"empty data",
			make(url.Values),
			TestTarget{
				String: "banana", Ptr: nil, PtrDefault: ptr("orange"),
				Uint64: 7, Int: 42, Int64: 66, Float64: 3.14159,
				Bool: true, Time: time.Now().AddDate(-1, 2, -10), Duration: time.Second,
				TextUnmarshaler: "text_text", BinaryUnmarshaler: []byte("data_binary"),
			},
		},
		{
			"non-empty data",
			map[string][]string{
				"aString":           {"apple"},
				"Ptr":               {"pear"},
				"PtrDefault":        {"watermelon"},
				"Uint":              {"1"},
				"Uint8":             {"2"},
				"Uint16":            {"3"},
				"Uint32":            {"4"},
				"aUint64":           {"5"},
				"aInt":              {"6"},
				"Int8":              {"7"},
				"Int16":             {"8"},
				"Int32":             {"9"},
				"aInt64":            {"10"},
				"Float32":           {"11.11"},
				"aFloat64":          {"12.12"},
				"aBool":             {"false"},
				"Skip":              {"whatever"},
				"-":                 {"not skipped"},
				"Time":              {"now+2y"},
				"Duration":          {"5h"},
				"aMap":              {"key1:value1", "key2:value2"},
				"items":             {"item1", "item2"},
				"Byte":              {"7"},
				"TextUnmarshaler":   {"a"},
				"BinaryUnmarshaler": {"b"},
			},
			TestTarget{
				String: "apple", Ptr: ptr("pear"), PtrDefault: ptr("watermelon"),
				Uint: 1, Uint8: 2, Uint16: 3, Uint32: 4, Uint64: 5, Int: 6,
				Int8: 7, Int16: 8, Int32: 9, Int64: 10, Float32: 11.11, Float64: 12.12, Bool: false,
				Skip: "", NotSkipped: "not skipped",
				Time:     time.Now().AddDate(2, 0, 0),
				Duration: time.Hour * 5,
				Map:      map[string]string{"key1": "value1", "key2": "value2"},
				Anon: struct {
					Slice []string "urlvalue:\"items\""
				}{Slice: []string{"item1", "item2"}},
				Embed:           Embed{Byte: 7},
				TextUnmarshaler: "text_a", BinaryUnmarshaler: []byte("data_b"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestTarget
			if err := urlvalues.Unmarshal(tt.in, &got); err != nil {
				t.Fatalf("urlvalues.Unmarshal(%v, %v) = %q, want <nil>", tt.in, &got, err)
			}

			if diff := cmp.Diff(got, tt.want, cmpopts.EquateApproxTime(time.Millisecond)); diff != "" {
				t.Errorf("urlvalues.Unmarshal(...) -got +want\n%s", diff)
			}
		})
	}
}

func TestUnmarshal_TimeLayout(t *testing.T) {

	// See time_test.go for an exhaustive list of time parsing tests.

	type Target struct {
		Alarm time.Time `urlvalue:"alarm,layout:Kitchen"`
	}

	in := url.Values{"alarm": {"4:16PM"}}
	want := Target{Alarm: time.Date(0, 1, 1, 16, 16, 0, 0, time.UTC)}

	var got Target
	if err := urlvalues.Unmarshal(in, &got); err != nil {
		t.Fatalf("urlvalues.Unmarshal(%v, %v) = %q, want <nil>", in, got, err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("urlvalues.Unmarshal(...) -got +want\n%s", diff)
	}
}

func TestUnmarshal_Error(t *testing.T) {
	t.Run("invalid default value", func(t *testing.T) {
		in := make(url.Values)
		var target struct {
			Value int `urlvalue:",default:notAnInt"`
		}
		got := urlvalues.Unmarshal(in, &target)

		var fieldErr *urlvalues.FieldError
		if !errors.As(got, &fieldErr) {
			t.Errorf("urlvalues.Unmarshal(%v, %v) = %v, want %q", in, &target, got, reflect.TypeOf(fieldErr).String())
		}

		var parseErr *urlvalues.ParseError
		if errors.As(got, &parseErr) {
			t.Errorf("should not return urlvalues.ParseError for invalid default value, got %v", parseErr)
		}
	})

	t.Run("unparseable input", func(t *testing.T) {
		in := url.Values{"Value": {"notAnInt"}}
		var target struct {
			Value int
		}
		got := urlvalues.Unmarshal(in, &target)

		var parseErr *urlvalues.ParseError
		if !errors.As(got, &parseErr) {
			t.Errorf("urlvalues.Unmarshal(%v, %v) = %v, want %q", in, &target, got, reflect.TypeOf(parseErr).String())
		}

		var fieldErr *urlvalues.FieldError
		if !errors.As(got, &fieldErr) {
			t.Error("urlvalues.ParseError should wrap urlvalues.FieldError")
		}
	})
}

func TestUnmarshal_WithParseOptions(t *testing.T) {
	type Target struct {
		Items []string `urlvalue:"items,default:item1<<!>>item2<<!>>item3"`
	}

	tests := []struct {
		name  string
		delim string
		want  Target
	}{
		{
			"empty delim",
			"", // Defaults to ;
			Target{[]string{"item1<<!>>item2<<!>>item3"}},
		},
		{
			"custom delim",
			"<<!>>",
			Target{[]string{"item1", "item2", "item3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := make(url.Values)

			var got Target
			if err := urlvalues.Unmarshal(in, &got, urlvalues.WithDelimiter(tt.delim)); err != nil {
				t.Fatalf("urlvalues.Unmarshal(%v, %v) = %q, want <nil>", in, got, err)
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("urlvalues.Unmarshal(...) -got +want\n%s", diff)
			}
		})
	}
}

func TestUnmarshal_Validation(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		in := make(url.Values)
		target := &struct{ SomeField string }{}
		if err := urlvalues.Unmarshal(in, target); err != nil {
			t.Errorf("urlvalues.Unmarshal(%v, %v) = %q, want <nil>", in, target, err)
		}
	})

	t.Run("not a struct pointer", func(t *testing.T) {
		tests := []struct {
			name   string
			target any
		}{
			{"string", ""},
			{"float64", 0.0},
			{"int", 0},
			{"struct{}", struct{}{}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				in := make(url.Values)
				got := urlvalues.Unmarshal(in, tt.target)
				want := urlvalues.ErrInvalidStruct
				if !errors.Is(got, want) {
					t.Errorf("urlvalues.Unmarshal(%v, %v) = %v, want %q", in, tt.target, got, want)
				}
			})
		}
	})

	t.Run("no fields", func(t *testing.T) {
		in := make(url.Values)
		target := &struct{}{}
		if err := urlvalues.Unmarshal(in, target); err == nil {
			t.Errorf("urlvalues.Unmarshal(%v, %v) = <nil>, want error", in, target)
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}
