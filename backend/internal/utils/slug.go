package utils

import "strings"

// GenerateSlug creates a consistent, file-system safe slug from an IPO name
func GenerateSlug(name string) string {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, "/", "-")
	return slug
}
