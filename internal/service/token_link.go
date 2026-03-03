package service

import (
	"fmt"
	"net/url"
	"strings"
)

func buildTokenURL(baseURL string, token string, linkType string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("invalid %s base URL: %w", linkType, err)
	}

	query := parsedURL.Query()
	query.Set("token", token)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}
