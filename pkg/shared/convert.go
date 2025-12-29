// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import "fmt"

// ConvertMapToStrings converts a map[string]any to map[string]string.
// This is useful for environment variables and secrets which must be strings.
// Supports string, number (int, int64, float64), and boolean values.
func ConvertMapToStrings(m map[string]any) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = convertValueToString(v)
	}
	return result
}

// convertValueToString converts any value to its string representation
func convertValueToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// JSON numbers are unmarshaled as float64
		// Check if it's an integer value
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		// Fallback: convert to string using fmt
		return fmt.Sprintf("%v", v)
	}
}
