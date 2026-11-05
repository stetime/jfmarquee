package jellyfin

import "fmt"

// HTTPError reports a non-200 response from Jellyfin.
type HTTPError struct {
	Status int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("jellyfin returned %d", e.Status)
}
