module github.com/gilby125/google-flights-api

go 1.20

require (
	github.com/anyascii/go v0.3.2
	github.com/browserutils/kooky v0.2.1-0.20240119192416-d4f81abd0200
	github.com/go-test/deep v1.1.0
	github.com/hashicorp/go-retryablehttp v0.7.4
	github.com/krisukox/google-flights-api v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.13.0
	google.golang.org/protobuf v1.31.0
)

// Remove the direct requirement for krisukox/google-flights-api
// since we're replacing it with our local module

replace github.com/krisukox/google-flights-api => ./

require github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
