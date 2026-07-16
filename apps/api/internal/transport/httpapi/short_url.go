package httpapi

import "strings"

func buildShortURL(baseURL, shortCode string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" || shortCode == "" {
		return ""
	}

	return baseURL + "/" + shortCode
}
