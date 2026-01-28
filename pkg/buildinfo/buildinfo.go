package buildinfo

import (
	"fmt"
	"os"
	"strings"
)

// These are intended to be set via -ldflags at build time.
// Example:
// go build -ldflags "-X github.com/gilby125/google-flights-api/pkg/buildinfo.Version=v1.2.3 -X github.com/gilby125/google-flights-api/pkg/buildinfo.Commit=$(git rev-parse --short HEAD) -X github.com/gilby125/google-flights-api/pkg/buildinfo.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Info() map[string]string {
	return map[string]string{
		"version": Version,
		"commit":  Commit,
		"date":    Date,
	}
}

// VersionString returns a user-facing build identifier for UI/logging.
// Environment variables override build-time ldflags:
// - APP_VERSION (e.g. "build-257")
// - APP_COMMIT  (e.g. short SHA)
// - APP_DATE    (RFC3339 timestamp)
func VersionString() string {
	v := strings.TrimSpace(os.Getenv("APP_VERSION"))
	if v == "" {
		v = Version
	}
	c := strings.TrimSpace(os.Getenv("APP_COMMIT"))
	if c == "" {
		c = Commit
	}
	if c != "" && c != "unknown" {
		return fmt.Sprintf("%s (%s)", v, c)
	}
	return v
}
