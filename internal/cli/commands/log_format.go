package commands

import "regexp"

// leadingTimestampPatterns are anchored to the start of the string.
// Order matters: more specific patterns (PHP bracket, ISO 8601) are tried first.
var leadingTimestampPatterns = []*regexp.Regexp{
	// PHP built-in server / Apache ErrorLog: [Sat Jun 13 09:14:57 2026] or [Sat Jun  3 09:14:57 2026]
	regexp.MustCompile(`^\[\w{3} \w{3}\s+\d{1,2} \d{2}:\d{2}:\d{2} \d{4}\]\s*`),
	// ISO 8601: 2026-06-13T09:14:57Z  2026-06-13T09:14:57.123Z  2026-06-13T09:14:57+00:00
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?\s*`),
	// Date with space separator: 2026-06-13 09:14:57  2026-06-13 09:14:57.123
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?\s*`),
	// nginx / Go stdlib: 2026/06/13 09:14:57
	regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\s*`),
}

// stripMessageTimestamp removes a leading timestamp from a log message, returning
// the remainder. If no known timestamp pattern is found at the start of the string,
// the original message is returned unchanged.
func stripMessageTimestamp(message string) string {
	for _, re := range leadingTimestampPatterns {
		if loc := re.FindStringIndex(message); loc != nil {
			return message[loc[1]:]
		}
	}
	return message
}
