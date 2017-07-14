package httputil

const (
	// HeaderAccept can be used by user agents to specify response media
	// types that are acceptable.
	HeaderAccept = "Accept"

	// HeaderAcceptCharset can be sent by a user agent to indicate what
	// charsets are acceptable in textual response content.
	HeaderAcceptCharset = "Accept-Charset"

	// HeaderAcceptEncoding can be used by user agents to indicate what
	// response content-codings are acceptable in the response.
	HeaderAcceptEncoding = "Accept-Encoding"

	// HeaderAcceptLanguage can be used by user agents to indicate the
	// set of natural languages that are preferred in the response.
	HeaderAcceptLanguage = "Accept-Language"

	// HeaderAuthorization allows a user agent to authenticate itself
	// with an origin server - usually, but not necessarily, after
	// receiving a 401 (Unauthorized) response.
	HeaderAuthorization = "Authorization"

	// HeaderContentEncoding indicates what content codings have been
	// applied to the representation, beyond those inherent in the media
	// type, and thus what decoding mechanisms have to be applied in
	// order to obtain data in the media type referenced by the
	// Content-Type header field.
	HeaderContentEncoding = "Content-Encoding"

	// HeaderContentLanguage describes the natural language(s) of the
	// intended audience for the representation.
	HeaderContentLanguage = "Content-Language"

	// HeaderContentLength can provide the anticipated size, as a decimal
	// number of octets, for a potential payload body.
	HeaderContentLength = "Content-Length"

	// HeaderContentLocation references a URI that can be used as an
	// identifier for a specific resource corresponding to the
	// representation in this message's payload.
	HeaderContentLocation = "Content-Location"

	// HeaderContentType indicates the media type of the associated
	// representation: either the representation enclosed in the message
	// payload or the selected representation, as determined by the
	// message semantics.
	HeaderContentType = "Content-Type"

	// HeaderLink provides a means for serialising one or more links in
	// HTTP headers.
	HeaderLink = "Link"

	// HeaderRetryAfter can be used to indicate how long the service is
	// expected to be unavailable to the requesting client.
	HeaderRetryAfter = "Retry-After"

	// HeaderUserAgent contains information about the UAC originating the
	// request.
	HeaderUserAgent = "User-Agent"
)
