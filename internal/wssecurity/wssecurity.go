// Package wssecurity implements WS-Security UsernameToken profile per
// WS-Security 1.1 with the ONVIF Profile digest rules:
//
//	digest = base64( SHA-1( nonce + created + password ) )
//
// Placeholder during M0; real implementation arrives in M3 (client outbound)
// and M7 (server inbound verify).
package wssecurity

import (
	"crypto/sha1"
	"encoding/base64"
	"time"
)

// Digest computes the ONVIF UsernameToken digest.
//
// Per the ONVIF Profile, the digest is Base64(SHA-1(nonce || created || password))
// where nonce is the raw bytes (not base64-encoded) prior to hashing.
func Digest(nonce []byte, created string, password string) string {
	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(created))
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Created formats a UTC timestamp in the xs:dateTime form required by
// WS-Security (ISO-8601, UTC, millisecond precision optional).
func Created(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}
