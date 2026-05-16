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
	id1 := tbl.Add(3.14)
	id2 := tbl.Add(3.14)

	if len(tbl.values) != 1 {
		t.Errorf("table should count unique values only")
	}

	val := tbl.GetAll()[0]
	if val.Identifier != id1 {
		t.Fatalf("unexpected identifier")
	}
	if val.Value != 3.14 {
		t.Fatalf("unexpected value")
	}

	if id1 != id2 {
		t.Fatalf("unexpected value")
	}
}
