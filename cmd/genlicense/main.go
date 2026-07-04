// genlicense mints early-tester license keys. PRIVATE TOOL — never ship.
//
// Usage:
//
//	go run ./cmd/genlicense -email alice@example.com
//	go run ./cmd/genlicense -email bob@x.com -days 60
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"relay/internal/license"
)

func main() {
	var (
		email = flag.String("email", "", "tester email (becomes the license subject)")
		days  = flag.Int("days", 30, "validity in days from today")
		key   = flag.String("key", "keys/signing_key.pem", "path to signing key")
	)
	flag.Parse()

	if strings.TrimSpace(*email) == "" {
		die("missing -email")
	}
	if *days <= 0 || *days > 3650 {
		die("-days must be between 1 and 3650")
	}

	priv, err := os.ReadFile(*key)
	if err != nil {
		die("read signing key: %v", err)
	}

	now := time.Now()
	tok, err := license.Sign(priv, license.Payload{
		Subject:  *email,
		IssuedAt: now.Format("2006-01-02"),
		Expires:  now.AddDate(0, 0, *days).Format("2006-01-02"),
	})
	if err != nil {
		die("sign: %v", err)
	}

	// Sanity-verify what we just minted (catches subtle key/embed mismatches).
	embeddedPub, err := os.ReadFile("internal/license/embedded_public_key.pem")
	if err == nil {
		if _, err := license.VerifyWithKey(tok, embeddedPub); err != nil {
			die("self-verify failed (embedded pubkey doesn't match signing key?): %v", err)
		}
	}

	exp := now.AddDate(0, 0, *days).Format("2006-01-02")
	fmt.Fprintf(os.Stderr, "✓ issued license for %s (expires %s, %d days)\n\n", *email, exp, *days)
	fmt.Println(tok)
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
	os.Exit(1)
}
