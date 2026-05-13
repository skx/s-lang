// Functions for our string table.

package compiler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

// StringEntry is used in populating our code,
// via the text/template engine.
type StringEntry struct {
	// Identifier is the ID we calculate for the given string
	Identifier string

	// Value is the value of the string
	Value string
}

// StringTable holds our state
type StringTable struct {
	values map[string]string
}

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

		v = strings.ReplaceAll(v, "\n", "\\n")
		v = strings.ReplaceAll(v, "\t", "\\t")
		v = strings.ReplaceAll(v, "\r", "\\r")

		res = append(res, StringEntry{
			Identifier: k,
			Value:      v,
		})
	}

	return res
}
