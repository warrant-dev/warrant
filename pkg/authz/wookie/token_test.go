package wookie

import (
	"reflect"
	"testing"
	"time"
)

func TestBasicSerialization(t *testing.T) {
	ts := time.UnixMicro(1687375083854)
	token := Token{
		ID:        25,
		Version:   1,
		Timestamp: ts,
	}
	tokenString := token.String()
	expectedString := "MjU7MTsxNjg3Mzc1MDgzODU0"
	if tokenString != expectedString {
		t.Fatalf("expected token string: %s, actual token string: %s", expectedString, tokenString)
	}

	deserializedToken, err := FromString(tokenString)
	if err != nil {
		t.Fatalf("unexpected error when deserializing token %v", err)
	}

	if !reflect.DeepEqual(token, deserializedToken) {
		t.Fatal("Deserialized token should be equal")
	}
}

func TestVariousInvalidTokenStrings(t *testing.T) {
	invalidString := ""
	_, err := FromString(invalidString)
	if err == nil {
		t.Fatalf("token string %s is invalid and should not deserialize", invalidString)
	}

	invalidString = "***"
	_, err = FromString(invalidString)
	if err == nil {
		t.Fatalf("token string %s is invalid and should not deserialize", invalidString)
	}

	invalidString = "MjU7MQ=="
	_, err = FromString(invalidString)
	if err == nil {
		t.Fatalf("token string %s is invalid and should not deserialize", invalidString)
	}

	invalidString = "aGk7MTsxNjg3Mzc1MDgzODU0"
	_, err = FromString(invalidString)
	if err == nil {
		t.Fatalf("token string %s is invalid and should not deserialize", invalidString)
	}

	invalidString = "MjU7eW91OzE2ODczNzUwODM4NTQ="
	_, err = FromString(invalidString)
	if err == nil {
		t.Fatalf("token string %s is invalid and should not deserialize", invalidString)
	}

	invalidString = "MjU7MTthc2Rm"
	_, err = FromString(invalidString)
	if err == nil {
		t.Fatalf("token string %s is invalid and should not deserialize", invalidString)
	}
}
