package apputils

import "testing"

func TestStrPointer(t *testing.T) {
	// Test case 1: Valid input string
	input1 := "Hello, World!"
	expected1 := "Hello, World!"
	result1 := StrPointer(input1)

	// Check the result
	if *result1 != expected1 {
		t.Errorf("StrPointer(%s) = %s; expected %s", input1, *result1, expected1)
	}

	// Test case 2: Empty input string
	input2 := ""
	expected2 := ""
	result2 := StrPointer(input2)

	// Check the result
	if *result2 != expected2 {
		t.Errorf("StrPointer(%s) = %s; expected %s", input2, *result2, expected2)
	}
}
