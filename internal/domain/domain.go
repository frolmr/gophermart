package domain

import "time"

const (
	TextContentType = "text/plain"
	HTMLContentType = "text/html"
	JSONContentType = "application/json"

	CompressFormat = "gzip"

	UserIDHeader = "X-User-ID"

	ZeroRetryAfter = 0 * time.Second

	ToSubunitDelimeter = 100
)
