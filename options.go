package urlvalues

// SetParseOptionFunc allows for overriding the parsing behaviour of URL values.
type SetParseOptionFunc func(*ParseOptions)

// WithDelimiter returns a SetParseOptionFunc that sets the delimiter used to
// convert slices and maps from and into their string representation.
func WithDelimiter(s string) SetParseOptionFunc {
	return func(o *ParseOptions) {
		o.delim = &s
	}
}

// ParseOptions holds all the options that allows for customizing the parsing
// behaviour when unmarshalling [url.Values].
type ParseOptions struct {
	// Delimiter used to convert slices and maps from and into their string
	// representaton.
	delim *string
}

// Delim returns the delimiter used to convert slices and maps from and into
// their string representation. Defaults to semicolon (;) if not set or set
// to the empty string.
func (o *ParseOptions) Delim() string {
	if o.delim != nil && *o.delim != "" {
		return *o.delim
	}
	return ";"
}
