package main

import (
	"strings"
)

// addLatestTagSuffix adds :latest to the image if it does not have a tag
func addLatestTagSuffix(image string) string {
	if !strings.Contains(image, ":") {
		return image + latestTagSuffix
	}
	return image
}
