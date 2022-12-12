package urlvalues

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalidStruct indicates that the Unmarshal target is not of correct type.
var ErrInvalidStruct = errors.New("urlvalues: target must be a struct pointer")

// ParseError occurs when a [url.Values] item failed to be parsed into a struct
// field's type.
type ParseError struct {
	// Name of struct field.
	FieldName string
	// Key into URL values.
	Key string

	fe *FieldError
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("error parsing value of %s: %s", e.Key, e.fe.err.Error())
}

// Unwrap returns the underlying [FieldError].
func (e *ParseError) Unwrap() error {
	return e.fe
}

// Unmarshal unmarshals data into the value pointed to by v. If v is nil or
// not a struct pointer, Unmarshal returns an [ErrInvalidStruct] error.
//
// Slices are decoded by splitting values by a delimiter and parsing each
// item individually. The delimiter defaults to semicolon (;), but can by
// customized by passing the [WithDelimiter] [SetParseOptionFunc]. Key-value pairs
// of maps are split using the same delimiter. Keys and their values are
// separated by a colon (:), with the key to the left and the value to the
// right of the colon.
//
// Fields with types implementing [encoding.TextUnmarshaller] and/or
// [encoding.BinaryUnmarshaller] will be decoded using those interfaces,
// respectively. If a type implements both interfaces, the
// [encoding.TextUnmarshaller] interface is used to decode the value.
//
// The decoding of each struct field can be customized by the name string
// stored under the "urlvalue" key in the struct field's tag. The name string
// acts as a key into data, possibly followed by a comma-separated list of
// options. The name may be empty, in which case the field name of the struct
// will act as as key into data in its stead.
//
// The "default" option allows for setting a default value on a field in case
// corresponding URL value is not present in data, or if the value is the zero
// value for the field's type.
//
// The "layout" option only applies to fields of type [time.Time] and allows for
// customizing how values should be parsed by providing layouts understood
// by [time.Parse]. See https://pkg.go.dev/time#pkg-constants for a complete list
// of the predefined layouts.
//
// As a special case, if the field tag is "-", the field is always omitted.
// Note that a field with name "-" can still be generated using the tag "-,".
//
// Examples of struct field tags and their meanings:
//
//	// Field defaults to 42.
//	Field int `urlvalue:"myName,default:42"`
//
//	// Field is parsed using the RFC850 layout.
//	Field time.Time `urlvalue:"myName,layout:RFC850"`
//
//	// Field is ignored by this package.
//	Field int `urlvalue:"-"`
//
//	// Field appears in URL values with key "-".
//	Field int `urlvalue:"-,"`
//
//	// Field is decoded as time.Now().AddDate(3, -4, 9).
//	Field time.Time `urlvalue:"myName,default:now+3y-4m+9d"`
//
// The parsing of [time.Time] is extended to support a "now" based parsing.
// It parses the value "now" to [time.Now]. Furthermore, it extends this
// syntax by allowing the consumer to subtract or add days (d), months (m)
// and years (y) to "now". This is done by prepending the date identifiers
// (d,m,y) with a minus (-) or plus (+) sign.
//
// Any error that occurs while processing struct fields results in a [FieldError].
// [ParseError] wraps around FieldError and is returned if any error occurs while
// parsing the [url.Values] that was passed into Unmarshal. ParseError is
// never returned from errors occuring while parsing default values.
func Unmarshal(data url.Values, v any, setParseOpts ...SetParseOptionFunc) error {
	pOpts := &ParseOptions{}
	for _, f := range setParseOpts {
		f(pOpts)
	}

	fields, err := extractFields(v)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return errors.New("urlvalues: no fields identified in target struct")
	}

	for _, field := range fields {
		field := field

		// Set any default value into the struct for this field.
		if field.options.defaultValue != "" {
			if err := processField(true, field.options.defaultValue, field.field, field.options, *pOpts); err != nil {
				return &FieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     field.options.defaultValue,
					err:       err,
				}
			}
		}

		// Extract access key into data for this field. Default to field name if
		// custom key not set in tags.
		key := field.options.key
		if key == "" {
			key = field.name
		}

		values, ok := data[key]
		if !ok || len(values) == 0 {
			continue
		}

		value := values[0]
		if len(values) > 1 {
			value = strings.Join(values, pOpts.Delim())
		}

		if err := processField(false, value, field.field, field.options, *pOpts); err != nil {
			return &ParseError{
				FieldName: field.name,
				Key:       key,
				fe: &FieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     value,
					err:       err,
				},
			}
		}
	}

	return nil
}
