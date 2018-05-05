package netutil

import (
	"crypto/tls"

	"github.com/ybubnov/memhashd/system/log"
)

// TLSConfig builds a TLS configuration based on the given certificate
// and key files. Function returns nil when either certificate or key
// is an empty string.
func TLSConfig(certFile, keyFile string) *tls.Config {
	if keyFile == "" || certFile == "" {
		return nil
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.ErrorLogf("net/TLS_CONFIG",
			"failed to load x509 key and centificate, %s", err)
		return nil
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
}
