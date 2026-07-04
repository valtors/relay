package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	outDir := flag.String("out", "keys", "output directory for keypair")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o700); err != nil {
		die(err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		die(err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		die(err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		die(err)
	}

	privPath := filepath.Join(*outDir, "signing_key.pem")
	pubPath := filepath.Join(*outDir, "public_key.pem")

	if _, err := os.Stat(privPath); err == nil {
		fmt.Fprintf(os.Stderr, "REFUSING TO OVERWRITE %s — keypair already exists.\n", privPath)
		fmt.Fprintln(os.Stderr, "Regenerating would invalidate every license you've issued. Delete the file manually if you really mean it.")
		os.Exit(1)
	}

	writePEM(privPath, "PRIVATE KEY", privBytes, 0o600)
	writePEM(pubPath, "PUBLIC KEY", pubBytes, 0o644)

	fmt.Printf("✓ wrote %s (KEEP SECRET — gitignored)\n", privPath)
	fmt.Printf("✓ wrote %s (embed into binary)\n", pubPath)
}

func writePEM(path, kind string, der []byte, mode os.FileMode) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		die(err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: kind, Bytes: der}); err != nil {
		die(err)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
