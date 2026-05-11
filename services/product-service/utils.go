package main

import (
	"os"
	"strings"
)

func resolveImageURL(imageURL string) string {
	if imageURL == "" {
		return imageURL
	}

	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		return imageURL
	}

	baseURL := strings.TrimSpace(os.Getenv("BASE_URL"))
	if baseURL == "" {
		return imageURL
	}

	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(imageURL, "/") {
		imageURL = "/" + imageURL
	}

	return baseURL + imageURL
}
