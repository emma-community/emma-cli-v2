// Package apierrors provides error formatting for emma API errors.
package apierrors

import (
	"errors"
	"fmt"
	"net/url"

	emma "github.com/emma-community/emma-go-sdk"
)

// Format returns a human-readable error for known emma API error types.
func Format(err error) error {
	if err == nil {
		return nil
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("network error: %w", urlErr)
	}

	var genericErr emma.GenericOpenAPIError
	if errors.As(err, &genericErr) {
		body := genericErr.Body()
		if len(body) > 0 {
			return fmt.Errorf("API error: %s", string(body))
		}
		return fmt.Errorf("API error: %s", genericErr.Error())
	}

	return err
}

// IsUnauthorized returns true if the error represents an unauthorized (401) error.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	var genericErr emma.GenericOpenAPIError
	if errors.As(err, &genericErr) {
		return genericErr.Error() == "401 Unauthorized"
	}
	return false
}

// IsNotFound returns true if the error represents a not-found (404) error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var genericErr emma.GenericOpenAPIError
	if errors.As(err, &genericErr) {
		return genericErr.Error() == "404 Not Found"
	}
	return false
}
