package license

import (
	"testing"
)

func TestVerifyString_Empty(t *testing.T) {
	_, err := VerifyString("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestVerifyString_Invalid(t *testing.T) {
	_, err := VerifyString("not-a-valid-license")
	if err == nil {
		t.Error("expected error for invalid license")
	}
}

func TestMissingError_Error(t *testing.T) {
	err := MissingError{}
	if err.Error() != "no license found" {
		t.Errorf("expected 'no license found', got %s", err.Error())
	}
}

func TestVerifyWithKey_Empty(t *testing.T) {
	_, err := VerifyWithKey("", []byte("fake-key"))
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestVerifyWithKey_InvalidKey(t *testing.T) {
	_, err := VerifyWithKey("some-license", []byte("not-a-valid-pem-key"))
	if err == nil {
		t.Error("expected error for invalid key")
	}
}
