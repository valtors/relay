// Package license verifies offline-signed early-tester licenses.
//
// License format:  RELAY-<base64url(payload-json)>.<base64url(ed25519-sig)>
//
// Payload is JSON: {"sub":"alice@x.com","iat":"2026-05-05","exp":"2026-06-04"}.
// The signature is over the raw payload bytes (NOT the base64 wrapping).
//
// Verification is purely offline against the embedded public key — no
// callout, no telemetry. A revocation list could be added later by
// embedding a slice of revoked subject IDs.
package license

import (
	"crypto/ed25519"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed embedded_public_key.pem
var embeddedPubKeyPEM []byte

const (
	prefix     = "RELAY-"
	dateLayout = "2006-01-02"

	// EnvVar is the canonical environment variable name. Stays consistent
	// across CLI/HTTP modes so users only have to learn one knob.
	EnvVar = "RELAY_LICENSE"

	// FormURL is shown in error messages. Replace once the public form
	// exists; the placeholder is intentionally obvious.
	FormURL = "https://tally.so/r/jaGYNa"
)

// Payload is the public claim set inside a license. Dates are
// YYYY-MM-DD strings to keep the encoded license short and human-readable.
type Payload struct {
	Subject  string `json:"sub"`
	IssuedAt string `json:"iat"`
	Expires  string `json:"exp"`
}

// Verify is the only function callers need. It pulls the license from the
// env var (preferred) or ~/.relay/license, parses + verifies
// the Ed25519 signature against the embedded public key, then checks
// expiry. Every failure path returns a user-actionable error.
func Verify() (*Payload, error) {
	raw, source, err := load()
	if err != nil {
		return nil, err
	}
	p, err := VerifyString(raw)
	if err != nil {
		return nil, fmt.Errorf("license from %s: %w", source, err)
	}
	return p, nil
}

// VerifyString verifies a license string in isolation. Exposed so the
// genlicense CLI can sanity-check what it just minted, and so tests can
// drive verification without touching the filesystem.
func VerifyString(raw string) (*Payload, error) {
	return verify(strings.TrimSpace(raw), embeddedPubKeyPEM)
}

// VerifyWithKey is the testing seam: same as VerifyString but takes the
// public key explicitly so tests can use a throwaway keypair without
// having to swap the embedded one.
func VerifyWithKey(raw string, pubKeyPEM []byte) (*Payload, error) {
	return verify(strings.TrimSpace(raw), pubKeyPEM)
}

func verify(raw string, pubKeyPEM []byte) (*Payload, error) {
	if raw == "" {
		return nil, errors.New("empty license string")
	}
	if !strings.HasPrefix(raw, prefix) {
		return nil, fmt.Errorf("license must start with %q", prefix)
	}
	body := strings.TrimPrefix(raw, prefix)

	parts := strings.SplitN(body, ".", 2)
	if len(parts) != 2 {
		return nil, errors.New("malformed license: expected <payload>.<signature>")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	pub, err := parsePublicKey(pubKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	if !ed25519.Verify(pub, payloadJSON, sig) {
		return nil, errors.New("signature verification failed (license has been tampered with or was not issued by us)")
	}

	var p Payload
	if err := json.Unmarshal(payloadJSON, &p); err != nil {
		return nil, fmt.Errorf("decode payload json: %w", err)
	}
	if p.Subject == "" || p.IssuedAt == "" || p.Expires == "" {
		return nil, errors.New("license payload missing required fields")
	}

	exp, err := time.Parse(dateLayout, p.Expires)
	if err != nil {
		return nil, fmt.Errorf("parse expiry %q: %w", p.Expires, err)
	}
	// Inclusive end-of-day so a license dated 2026-06-04 is valid until
	// 2026-06-04 23:59:59 in the user's local timezone.
	if time.Now().After(exp.Add(24 * time.Hour)) {
		return nil, fmt.Errorf("license expired on %s (subject: %s)", p.Expires, p.Subject)
	}
	return &p, nil
}

// load returns (raw, source, error) — source is a human-readable string
// like "$RELAY_LICENSE" or the file path, used to make error messages
// tell the user *where* to fix the problem.
func load() (string, string, error) {
	if v := strings.TrimSpace(os.Getenv(EnvVar)); v != "" {
		return v, "$" + EnvVar, nil
	}
	home, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(home, ".relay", "license")
		if b, err := os.ReadFile(path); err == nil {
			return string(b), path, nil
		}
	}
	return "", "", &MissingError{}
}

// MissingError is its own type so main.go can format it specially —
// missing license is "you haven't installed yet", invalid is "what you
// installed is broken", they deserve different copy.
type MissingError struct{}

func (MissingError) Error() string {
	return "no license found"
}

// FriendlyMessage returns a multi-line block suitable for printing to
// stderr when verification fails. The exact wording is centralised here
// so error copy stays consistent across stdio and HTTP transports.
func FriendlyMessage(err error) string {
	var sb strings.Builder
	sb.WriteString("\n┌── relay ───────────────────────────────────┐\n")

	var missing *MissingError
	if errors.As(err, &missing) {
		sb.WriteString("│ No license key found.                                │\n")
		sb.WriteString("│                                                      │\n")
		sb.WriteString("│ This is a closed-beta build. To request a key:       │\n")
		fmt.Fprintf(&sb, "│   %-50s │\n", FormURL)
		sb.WriteString("│                                                      │\n")
		sb.WriteString("│ Once you have one, install it via either:            │\n")
		fmt.Fprintf(&sb, "│   export %s=RELAY-...                     │\n", EnvVar)
		sb.WriteString("│   or write it to ~/.relay/license         │\n")
	} else {
		sb.WriteString("│ License rejected.                                    │\n")
		sb.WriteString("│                                                      │\n")
		fmt.Fprintf(&sb, "│ Reason: %-44s │\n", truncate(err.Error(), 44))
		sb.WriteString("│                                                      │\n")
		sb.WriteString("│ Need a fresh key? Request one at:                    │\n")
		fmt.Fprintf(&sb, "│   %-50s │\n", FormURL)
	}
	sb.WriteString("└──────────────────────────────────────────────────────┘\n")
	return sb.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func parsePublicKey(pemBytes []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block found")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	edPub, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected ed25519 public key, got %T", pub)
	}
	return edPub, nil
}

// Sign produces a license string. Lives here (not in cmd/genlicense) so
// tests can mint test licenses without depending on the CLI binary.
// Pass a private key in PKCS8 PEM form.
func Sign(privKeyPEM []byte, p Payload) (string, error) {
	if p.Subject == "" || p.IssuedAt == "" || p.Expires == "" {
		return "", errors.New("payload missing required fields")
	}
	if _, err := time.Parse(dateLayout, p.IssuedAt); err != nil {
		return "", fmt.Errorf("iat must be %s: %w", dateLayout, err)
	}
	if _, err := time.Parse(dateLayout, p.Expires); err != nil {
		return "", fmt.Errorf("exp must be %s: %w", dateLayout, err)
	}

	priv, err := parsePrivateKey(privKeyPEM)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, body)
	return prefix +
		base64.RawURLEncoding.EncodeToString(body) + "." +
		base64.RawURLEncoding.EncodeToString(sig), nil
}

func parsePrivateKey(pemBytes []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block in signing key")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	edPriv, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected ed25519 private key, got %T", priv)
	}
	return edPriv, nil
}
