package wssecurity

import (
	"crypto/sha1"
	"encoding/base64"
	"testing"
	"time"
)

func TestDigest_MatchesSpecFormula(t *testing.T) {
	nonce := []byte("0123456789ABCDEF")
	created := "2026-07-05T12:00:00Z"
	password := "secret"

	got := Digest(nonce, created, password)

	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(created))
	h.Write([]byte(password))
	want := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if got != want {
		t.Fatalf("digest mismatch: got %q want %q", got, want)
	}
}

func TestCreated_UTC(t *testing.T) {
	c := Created(parseTime(t, "2026-07-05T12:00:00+02:00"))
	if c != "2026-07-05T10:00:00Z" {
		t.Fatalf("got %q want UTC normalised timestamp", c)
	}
}

func parseTime(t *testing.T, s string) (tm time.Time) {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return
}