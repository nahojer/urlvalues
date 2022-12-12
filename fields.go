package urlvalues

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FieldError occurs when an error occurs updating an individual field in the
// provided struct value.
type FieldError struct {
	fieldName string
	typeName  string
	value     string
	err       error
}

func (err *FieldError) Error() string {
	return fmt.Sprintf("urlvalues: error assigning to field %s: converting '%s' to type %s. details: %s", err.fieldName, err.value, err.typeName, err.err)
}

// field maintains information about a field in the target struct.
type field struct {
	name    string
	field   reflect.Value
	options fieldOptions
}

// fieldOptions maintain options for a given field.
type fieldOptions struct {
	key          string
	defaultValue string
	layout       string
}

func extractFields(target any) ([]field, error) {
	strct := reflect.ValueOf(target)
	if strct.Kind() != reflect.Ptr {
		return nil, ErrInvalidStruct
	}
	strct = strct.Elem()
	if strct.Kind() != reflect.Struct {
		return nil, ErrInvalidStruct
	}

	var fields []field
	for i := 0; i < strct.NumField(); i++ {
		f := strct.Field(i)
		strctField := strct.Type().Field(i)

		// Get the urlvalue tags associated with this field (if any).
		fieldTags := strctField.Tag.Get("urlvalue")

		// If it's ignored or can't be set, move on.
		if !f.CanSet() || fieldTags == "-" {
			continue
		}

		fieldName := strctField.Name

		fieldOpts, err := parseTag(fieldTags)
		if err != nil {
			return nil, fmt.Errorf("urlvalues: parsing tags for field %s: %w", fieldName, err)
		}

		// Drill down through pointers until we bottom out at type or nil.
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				// It's not a struct so leave it alone.
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}

				// It is a struct so zero it out.
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		switch {
		// If we found a struct that can't deserialize itself, drill down, appending
		// fields as we go.
		case f.Kind() == reflect.Struct && textUnmarshaler(f) == nil && binaryUnmarshaler(f) == nil:
			embeddedPtr := f.Addr().Interface()
			innerFields, err := extractFields(embeddedPtr)
			if err != nil {
				return nil, fmt.Errorf("urlvalues: %w", err)
			}
			fields = append(fields, innerFields...)
		default:
			fields = append(fields, field{
				name:    fieldName,
				field:   f,
				options: fieldOpts,
			})
		}
	}

	return fields, nil
}

func parseTag(tagStr string) (fieldOptions, error) {
	if tagStr == "" {
		return fieldOptions{}, nil
	}

	var fOpts fieldOptions

	tagParts := strings.Split(tagStr, ",")
	for i, tagPart := range tagParts {
		vals := strings.SplitN(tagPart, ":", 2)
		tagProp := strings.TrimSpace(vals[0])

		switch len(vals) {
		case 1:
			switch tagProp {
			default:
				if i == 0 {
					fOpts.key = tagProp
				}
			}
		case 2:
			tagPropVal := strings.TrimSpace(vals[1])
			if tagPropVal == "" {
				return fOpts, fmt.Errorf("tag %q missing a value", tagProp)
			}
			switch tagProp {
			case "default":
				fOpts.defaultValue = tagPropVal
			case "layout":
				fOpts.layout = tagPropVal
			}
		}
	}

	return fOpts, nil
}

func processField(settingDefault bool, value string, field reflect.Value, fOpts fieldOptions, pOpts ParseOptions) error {
	typ := field.Type()

	// Extend time.Time parsing to accept custom layouts and our own "now" based
	// parsing.
	if typ.PkgPath() == "time" && typ.Name() == "Time" {
		tim, err := parseTime(fOpts.layout, value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(tim))
		return nil
	}

	// Types implementing encoding.TextUnmarshaler.
	if t := textUnmarshaler(field); t != nil {
		return t.UnmarshalText([]byte(value))
	}

	// Types implementing encoding.BinaryUnmarshaler.
	if b := binaryUnmarshaler(field); b != nil {
		return b.UnmarshalBinary([]byte(value))
	}

	// Dereference pointer.
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if field.IsNil() {
			field.Set(reflect.New(typ))
		}
		field = field.Elem()
	}

	// We don't want a default value to override a proper setting.
	if settingDefault && !field.IsZero() {
		return nil
	}

	// Builtin types.
	switch typ.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)
		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, typ.Bits())
		}
		if err != nil {
			return err
		}
		field.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			return err
		}
		field.SetUint(val)

	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(val)

	case reflect.Slice:
		vals := strings.Split(value, pOpts.Delim())
		sl := reflect.MakeSlice(typ, len(vals), len(vals))
		for i, val := range vals {
			err := processField(false, val, sl.Index(i), fOpts, pOpts)
			if err != nil {
				return err
			}
		}
		field.Set(sl)

	case reflect.Map:
		mp := reflect.MakeMap(typ)
		if len(strings.TrimSpace(value)) != 0 {
			pairs := strings.Split(value, pOpts.Delim())
			for _, pair := range pairs {
				kvpair := strings.Split(pair, ":")
				if len(kvpair) != 2 {
					return fmt.Errorf("invalid map item: %q", pair)
				}
				k := reflect.New(typ.Key()).Elem()
				err := processField(false, kvpair[0], k, fOpts, pOpts)
				if err != nil {
					return err
				}
				v := reflect.New(typ.Elem()).Elem()
				err = processField(false, kvpair[1], v, fOpts, pOpts)
				if err != nil {
					return err
				}
				mp.SetMapIndex(k, v)
			}
		}
		field.Set(mp)
	}

	return nil
}

func textUnmarshaler(field reflect.Value) (t encoding.TextUnmarshaler) {
	interfaceFrom(field, func(v any, ok *bool) {
		t, *ok = v.(encoding.TextUnmarshaler)
	})
	return t
}

func binaryUnmarshaler(field reflect.Value) (b encoding.BinaryUnmarshaler) {
	interfaceFrom(field, func(v any, ok *bool) {
		b, *ok = v.(encoding.BinaryUnmarshaler)
	})
	return b
}

func interfaceFrom(field reflect.Value, fn func(any, *bool)) {
	if !field.CanInterface() {
		return
	}

	var ok bool
	fn(field.Interface(), &ok)
	if !ok && field.CanAddr() {
		fn(field.Addr().Interface(), &ok)
	}
}
