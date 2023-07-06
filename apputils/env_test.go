package apputils

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Set up test environment
	key := "TEST_KEY"
	fallback := "fallback_value"
	expectedValue := "test_value"
	os.Setenv(key, expectedValue)

	// Call the function
	result := GetEnv(key, fallback)

	// Check the result
	if result != expectedValue {
		t.Errorf("GetEnv(%s, %s) = %s; expected %s", key, fallback, result, expectedValue)
	}

	// Clean up test environment
	os.Unsetenv(key)

	// Call the function with fallback value
	result = GetEnv("NON_EXISTENT_KEY", fallback)

	// Check the result with fallback value
	if result != fallback {
		t.Errorf("GetEnv(%s, %s) = %s; expected %s", "NON_EXISTENT_KEY", fallback, result, fallback)
	}
}
