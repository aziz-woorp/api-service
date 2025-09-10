// Package utils provides utility functions for CSAT operations.
package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateCSATType validates that a CSAT type follows snake_case naming convention.
// Rules:
// - Cannot be empty
// - Cannot contain spaces
// - Must be lowercase
// - Can only contain lowercase letters, numbers, and underscores
func ValidateCSATType(csatType string) error {
	// Check for empty
	if strings.TrimSpace(csatType) == "" {
		return fmt.Errorf("CSAT type cannot be empty")
	}
	
	// Check for spaces
	if strings.Contains(csatType, " ") {
		return fmt.Errorf("CSAT type cannot contain spaces")
	}
	
	// Check for uppercase (should be lowercase)
	if csatType != strings.ToLower(csatType) {
		return fmt.Errorf("CSAT type must be lowercase")
	}
	
	// Check for valid characters (letters, numbers, underscores)
	matched, _ := regexp.MatchString("^[a-z0-9_]+$", csatType)
	if !matched {
		return fmt.Errorf("CSAT type can only contain lowercase letters, numbers, and underscores")
	}
	
	return nil
}
