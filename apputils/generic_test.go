package apputils

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	// Test case 1: Mapping integers to strings
	input1 := []int{1, 2, 3, 4, 5}
	expected1 := []string{"1", "2", "3", "4", "5"}
	result1 := Map(input1, func(n int) string {
		return strconv.Itoa(n)
	})

	// Check the result
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Map(%v) = %v; expected %v", input1, result1, expected1)
	}

	// Test case 2: Mapping strings to uppercase
	input2 := []string{"apple", "banana", "cherry"}
	expected2 := []string{"APPLE", "BANANA", "CHERRY"}
	result2 := Map(input2, strings.ToUpper)

	// Check the result
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Map(%v) = %v; expected %v", input2, result2, expected2)
	}

	// Test case 3: Mapping empty slice
	var input3 []int
	input3 = []int{}
	var expected3 []string
	expected3 = []string{}
	result3 := Map(input3, func(n int) string {
		return strconv.Itoa(n)
	})

	// Check the result
	if !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Map(%v) = %v; expected %v", input3, result3, expected3)
	}
}
