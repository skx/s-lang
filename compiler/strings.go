// Functions for our string table.

package compiler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

// StringEntry is used in populating our code, via the text/template engine.
type StringEntry struct {
	// Identifier is the ID we calculate for the given string.
	Identifier string

	// Length is the length of the string, not including the trailing
	// null-terminator.
	Length int

	// Value is the literal value of the string.
	Value string

	// Encoded is the hex-encoded variant of the string.
	Encoded string
}

// StringTable holds our state.
type StringTable struct {
	// values holds known-strings, and their details.
	values map[string]string
}

// NewStringTable is our constructor.
func NewStringTable() *StringTable {
	return &StringTable{
		values: make(map[string]string),
	}
}

// Add inserts a new entry into our string table,
// generating and returning a safe identifier for it.
func (st *StringTable) Add(str string) string {

	// Create ID
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))
	id := fmt.Sprintf("str_%s", sha)

	// save it
	st.values[id] = str

	// return it
	return (id)
}

// GetAll returns our string-table as an array of
// objects, suitable for using in a text/template
// file.
func (st *StringTable) GetAll() []StringEntry {
	res := []StringEntry{}

	for k, v := range st.values {

		// Convert "Steve" into "s, t, e, v, e" (hex encoded)
		var b strings.Builder
		comma := false
		for _, c := range []byte(v) {
			if !comma {
				fmt.Fprintf(&b, "0x%02x", c)
				comma = true
			} else {
				fmt.Fprintf(&b, ", 0x%02x", c)
			}
		}

		res = append(res, StringEntry{
			Identifier: k,
			Length:     len(v) + 1,
			Value:      v,
			Encoded:    b.String(),
		})
	}

	return res
}
