// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
