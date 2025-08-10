package f1

// TruncateString truncates a string to a maximum length and adds "..." if needed
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}