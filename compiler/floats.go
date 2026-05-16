// Functions for our float table.

package compiler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// FloatEntry is used in populating our code,
// via the text/template engine.
type FloatEntry struct {
	// Identifier is the ID we calculate for the given string
	Identifier string

	// Value is the float value.
	Value float64
}

// FloatTable holds our state.
type FloatTable struct {
	values map[string]float64
}

// NewFloatTable is our constructor.
func NewFloatTable() *FloatTable {
	return &FloatTable{
		values: make(map[string]float64),
	}
}

// Add inserts a new entry into our table, generating and
// returning a safe identifier for it.
func (ft *FloatTable) Add(f float64) string {

	// Create ID
	hasher := sha1.New()
	hasher.Write([]byte(fmt.Sprintf("%f", f)))
	sha := hex.EncodeToString(hasher.Sum(nil))
	id := fmt.Sprintf("float_%s", sha)

	// save it
	ft.values[id] = f

	// return it
	return (id)
}

// GetAll returns our string-table as an array of
// objects, suitable for using in a text/template
// file.
func (ft *FloatTable) GetAll() []FloatEntry {
	res := []FloatEntry{}

	for k, v := range ft.values {

		res = append(res, FloatEntry{
			Identifier: k,
			Value:      v,
		})
	}

	return res
}
