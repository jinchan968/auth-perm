package utils

import "strings"

// EscapeILIKE escapes SQL ILIKE special characters (% and _) in user input.
func EscapeILIKE(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// ILIKEPattern wraps a keyword in % wildcards with escaped special chars.
func ILIKEPattern(keyword string) string {
	return "%" + EscapeILIKE(keyword) + "%"
}
