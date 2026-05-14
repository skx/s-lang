package compiler

import (
	"testing"
)

// TestFloatTable makes trivial tests of our float-table.
func TestFloatTable(t *testing.T) {

	tbl := NewFloatTable()

	if len(tbl.values) != 0 {
		t.Errorf("new table is not empty")
	}

	// add the same entry multiple times
	id := tbl.Add(3.14)
	id = tbl.Add(3.14)

	if len(tbl.values) != 1 {
		t.Errorf("table should count unique values only")
	}

	val := tbl.GetAll()[0]
	if val.Identifier != id {
		t.Fatalf("unexpected identifier")
	}
	if val.Value != 3.14 {
		t.Fatalf("unexpected value")
	}
}
