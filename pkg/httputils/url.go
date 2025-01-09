package httputils

import (
	"fmt"
	"net/url"
)

func GetURL(baseURL string, queryParams map[string]string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	query := url.Values{}
	for key, value := range queryParams {
		query.Set(key, value)
	}

	parsedURL.RawQuery = query.Encode()

	finalURL := parsedURL.String()

	return finalURL, nil
}
