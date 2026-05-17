package provider

import (
	"errors"
)

func isStatus(err error, status int) bool {
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == status
	}
	return false
}
