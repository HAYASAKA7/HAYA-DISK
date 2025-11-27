package utils

import (
	"regexp"
	"strings"
)

// PhoneRegionInfo contains information about a phone region
type PhoneRegionInfo struct {
	Name      string // Full country/region name
	Code      string // International dialing code
	Pattern   string // Regex pattern for validation
	Example   string // Example phone number
	MinLength int    // Minimum digits
	MaxLength int    // Maximum digits
}

// PhoneRegions contains phone validation rules for different regions
var PhoneRegions = map[string]PhoneRegionInfo{
	"CN": {
		Name:      "China",
		Code:      "+86",
		Pattern:   `^1[3-9]\d{9}$`,
		Example:   "13812345678",
		MinLength: 11,
		MaxLength: 11,
	},
	"US": {
		Name:      "United States",
		Code:      "+1",
		Pattern:   `^\d{10}$`,
		Example:   "2025551234",
		MinLength: 10,
		MaxLength: 10,
	},
	"UK": {
		Name:      "United Kingdom",
		Code:      "+44",
		Pattern:   `^7\d{9}$`,
		Example:   "7911123456",
		MinLength: 10,
		MaxLength: 10,
	},
	"JP": {
		Name:      "Japan",
		Code:      "+81",
		Pattern:   `^[789]0\d{8}$`,
		Example:   "9012345678",
		MinLength: 10,
		MaxLength: 10,
	},
	"KR": {
		Name:      "South Korea",
		Code:      "+82",
		Pattern:   `^01[0-9]\d{7,8}$`,
		Example:   "01012345678",
		MinLength: 10,
		MaxLength: 11,
	},
	"TW": {
		Name:      "Taiwan",
		Code:      "+886",
		Pattern:   `^9\d{8}$`,
		Example:   "912345678",
		MinLength: 9,
		MaxLength: 9,
	},
	"HK": {
		Name:      "Hong Kong",
		Code:      "+852",
		Pattern:   `^[5-9]\d{7}$`,
		Example:   "51234567",
		MinLength: 8,
		MaxLength: 8,
	},
	"SG": {
		Name:      "Singapore",
		Code:      "+65",
		Pattern:   `^[89]\d{7}$`,
		Example:   "81234567",
		MinLength: 8,
		MaxLength: 8,
	},
	"AU": {
		Name:      "Australia",
		Code:      "+61",
		Pattern:   `^4\d{8}$`,
		Example:   "412345678",
		MinLength: 9,
		MaxLength: 9,
	},
	"DE": {
		Name:      "Germany",
		Code:      "+49",
		Pattern:   `^1[5-7]\d{8,9}$`,
		Example:   "15123456789",
		MinLength: 10,
		MaxLength: 11,
	},
	"FR": {
		Name:      "France",
		Code:      "+33",
		Pattern:   `^[67]\d{8}$`,
		Example:   "612345678",
		MinLength: 9,
		MaxLength: 9,
	},
	"IN": {
		Name:      "India",
		Code:      "+91",
		Pattern:   `^[6-9]\d{9}$`,
		Example:   "9123456789",
		MinLength: 10,
		MaxLength: 10,
	},
	"CA": {
		Name:      "Canada",
		Code:      "+1",
		Pattern:   `^\d{10}$`,
		Example:   "4165551234",
		MinLength: 10,
		MaxLength: 10,
	},
	"RU": {
		Name:      "Russia",
		Code:      "+7",
		Pattern:   `^9\d{9}$`,
		Example:   "9123456789",
		MinLength: 10,
		MaxLength: 10,
	},
	"BR": {
		Name:      "Brazil",
		Code:      "+55",
		Pattern:   `^[1-9]{2}9?\d{8}$`,
		Example:   "11912345678",
		MinLength: 10,
		MaxLength: 11,
	},
	"MX": {
		Name:      "Mexico",
		Code:      "+52",
		Pattern:   `^[1-9]\d{9}$`,
		Example:   "5512345678",
		MinLength: 10,
		MaxLength: 10,
	},
}

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address format
func ValidateEmail(email string) bool {
	if email == "" {
		return true // Empty email is allowed (optional field)
	}
	email = strings.TrimSpace(email)
	if len(email) > 254 { // Max email length per RFC
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidatePhone validates a phone number based on the region
func ValidatePhone(phone string, region string) bool {
	if phone == "" {
		return true // Empty phone is allowed (optional field)
	}

	// Clean the phone number - remove spaces, dashes, parentheses
	phone = CleanPhoneNumber(phone)

	regionInfo, exists := PhoneRegions[region]
	if !exists {
		// If region not found, do basic validation (only digits, reasonable length)
		return regexp.MustCompile(`^\d{6,15}$`).MatchString(phone)
	}

	regex := regexp.MustCompile(regionInfo.Pattern)
	return regex.MatchString(phone)
}

// CleanPhoneNumber removes common formatting characters from phone number
func CleanPhoneNumber(phone string) string {
	// Remove spaces, dashes, parentheses, dots
	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"(", "",
		")", "",
		".", "",
		"+", "",
	)
	return replacer.Replace(phone)
}

// GetPhoneRegionInfo returns the region info for a given region code
func GetPhoneRegionInfo(region string) (PhoneRegionInfo, bool) {
	info, exists := PhoneRegions[region]
	return info, exists
}

// GetAllPhoneRegions returns all available phone regions
func GetAllPhoneRegions() map[string]PhoneRegionInfo {
	return PhoneRegions
}

// GetPhoneValidationError returns a user-friendly error message for invalid phone
func GetPhoneValidationError(region string) string {
	regionInfo, exists := PhoneRegions[region]
	if !exists {
		return "Invalid phone number format. Please enter 6-15 digits."
	}

	return "Invalid phone number format for " + regionInfo.Name +
		". Expected " + formatLengthDescription(regionInfo.MinLength, regionInfo.MaxLength) +
		" digits. Example: " + regionInfo.Example
}

// formatLengthDescription returns a human-readable length description
func formatLengthDescription(min, max int) string {
	if min == max {
		return string(rune('0'+min/10)) + string(rune('0'+min%10))
	}
	return string(rune('0'+min/10)) + string(rune('0'+min%10)) + "-" +
		string(rune('0'+max/10)) + string(rune('0'+max%10))
}

// ValidationResult holds the result of a validation check
type ValidationResult struct {
	Valid   bool
	Field   string
	Message string
}

// ValidateRegistration validates all registration fields
func ValidateRegistration(email, phone, phoneRegion string) []ValidationResult {
	var results []ValidationResult

	// Validate email if provided
	if email != "" && !ValidateEmail(email) {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "email",
			Message: "Invalid email format. Please enter a valid email address.",
		})
	}

	// Validate phone if provided
	if phone != "" && !ValidatePhone(phone, phoneRegion) {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "phone",
			Message: GetPhoneValidationError(phoneRegion),
		})
	}

	return results
}
