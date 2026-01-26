package buildinfo

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
