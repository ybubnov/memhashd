package httputil

const (
	// TypeAny is Accept header of the request that accepts data in any
	// format.
	TypeAny = "*/*"

	// TypeApplicationJSON is a JSON media type.
	TypeApplicationJSON = "application/json"

	// TypeApplicationYAML is an YAML media type.
	TypeApplicationYAML = "application/yaml"

	// TypeApplication is an application media type.
	TypeApplication = "application/*"

	// TypeTextHTML is a HTML media type.
	TypeTextHTML = "text/html"

	// TypeText is a text media type.
	TypeText = "text/*"
)
