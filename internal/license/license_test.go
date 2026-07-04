package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"testing"
	"time"
)

func genTestKeypair(t *testing.T) (privPEM, pubPEM []byte) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return
}

func today() string        { return time.Now().Format(dateLayout) }
func daysOut(d int) string { return time.Now().AddDate(0, 0, d).Format(dateLayout) }

func TestSignAndVerifyRoundTrip(t *testing.T) {
	priv, pub := genTestKeypair(t)
	tok, err := Sign(priv, Payload{
		Subject:  "alice@example.com",
		IssuedAt: today(),
		Expires:  daysOut(30),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tok, "RELAY-") {
		t.Fatalf("expected RELAY- prefix, got %q", tok[:8])
	}
	if !strings.Contains(tok, ".") {
		t.Fatal("expected '.' separator between payload and sig")
	}

	got, err := VerifyWithKey(tok, pub)
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != "alice@example.com" {
		t.Errorf("subject = %q, want alice@example.com", got.Subject)
	}
}

func TestVerify_RejectsTamperedPayload(t *testing.T) {
	priv, pub := genTestKeypair(t)
	tok, _ := Sign(priv, Payload{Subject: "a@b.com", IssuedAt: today(), Expires: daysOut(30)})

	parts := strings.SplitN(strings.TrimPrefix(tok, "RELAY-"), ".", 2)
	if len(parts) != 2 {
		t.Fatal("setup")
	}
	mid := len(parts[0]) / 2
	swap := byte('A')
	if parts[0][mid] == 'A' {
		swap = 'B'
	}
	tamperedPayload := parts[0][:mid] + string(swap) + parts[0][mid+1:]
	tampered := "RELAY-" + tamperedPayload + "." + parts[1]

	_, err := VerifyWithKey(tampered, pub)
	if err == nil {
		t.Fatal("expected verification to fail on tampered payload")
	}
}

func TestVerify_RejectsWrongPubKey(t *testing.T) {
	priv1, _ := genTestKeypair(t)
	_, pub2 := genTestKeypair(t)
	tok, _ := Sign(priv1, Payload{Subject: "a@b.com", IssuedAt: today(), Expires: daysOut(30)})

	_, err := VerifyWithKey(tok, pub2)
	if err == nil {
		t.Fatal("expected verification to fail when verifying with unrelated pubkey")
	}
}

func TestVerify_RejectsExpired(t *testing.T) {
	priv, pub := genTestKeypair(t)
	tok, _ := Sign(priv, Payload{
		Subject:  "a@b.com",
		IssuedAt: daysOut(-60),
		Expires:  daysOut(-2),
	})
	_, err := VerifyWithKey(tok, pub)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expiry error, got %v", err)
	}
}

func TestVerify_AcceptsExpiringToday(t *testing.T) {
	priv, pub := genTestKeypair(t)
	tok, _ := Sign(priv, Payload{Subject: "a@b.com", IssuedAt: daysOut(-30), Expires: today()})
	if _, err := VerifyWithKey(tok, pub); err != nil {
		t.Fatalf("expected today's expiry to still be valid: %v", err)
	}
}

func TestVerify_RejectsMalformedString(t *testing.T) {
	_, pub := genTestKeypair(t)
	cases := []string{"", "garbage", "RELAY-onlyhalf", "RELAY-not.base64!!"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, err := VerifyWithKey(c, pub); err == nil {
				t.Errorf("expected error for %q", c)
			}
		})
	}
}

func TestVerify_FromEnv(t *testing.T) {
	priv, pub := genTestKeypair(t)
	tok, _ := Sign(priv, Payload{Subject: "a@b.com", IssuedAt: today(), Expires: daysOut(30)})
	t.Setenv(EnvVar, tok)

	got, src, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if src != "$"+EnvVar {
		t.Errorf("source = %q, want $%s", src, EnvVar)
	}
	if _, err := VerifyWithKey(got, pub); err != nil {
		t.Fatal(err)
	}
}

func TestVerify_MissingErrorIsTyped(t *testing.T) {
	t.Setenv(EnvVar, "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())

	_, err := Verify()
	var miss *MissingError
	if !errors.As(err, &miss) {
		t.Fatalf("expected *MissingError, got %T (%v)", err, err)
	}
}

func TestFriendlyMessage_DistinguishesMissingVsInvalid(t *testing.T) {
	missing := FriendlyMessage(&MissingError{})
	if !strings.Contains(missing, "No license key found") {
		t.Errorf("missing message lacks expected copy: %s", missing)
	}
	invalid := FriendlyMessage(errors.New("signature verification failed"))
	if !strings.Contains(invalid, "License rejected") {
		t.Errorf("invalid message lacks expected copy: %s", invalid)
	}
}
