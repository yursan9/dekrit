package output

import (
	"encoding/json"
	"io"
)

// JSON writes v to w as indented JSON followed by a newline. HTML characters
// are not escaped.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
