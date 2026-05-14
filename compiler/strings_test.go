package compiler

import (
	"testing"
)

// TestStringTable makes trivial tests of our string-table
func TestStringTable(t *testing.T) {

	tbl := NewStringTable()

	if len(tbl.values) != 0 {
		t.Errorf("new table is not empty")
	}

	// add the same entry multiple times
	id := tbl.Add("Steve")
	id = tbl.Add("Steve")

	if len(tbl.values) != 1 {
		t.Errorf("table should count unique values only")
	}

	val := tbl.GetAll()[0]
	if val.Identifier != id {
		t.Fatalf("unexpected identifier")
	}
	if val.Value != "Steve" {
		t.Fatalf("unexpected value")
	}
}
