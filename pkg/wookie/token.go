// Copyright 2024 WorkOS, Inc.
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
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Token struct {
	ID        int64
	Version   int64
	Timestamp time.Time
}

// Get string representation of token (to set as header).
func (t Token) String() string {
	s := fmt.Sprintf("%d;%d;%d", t.ID, t.Version, t.Timestamp.UnixMicro())
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// De-serialize token from string (from header).
func FromString(wookieString string) (*Token, error) {
	if wookieString == "" {
		return nil, errors.New("empty wookie string")
	}
	decodedStr, err := base64.StdEncoding.DecodeString(wookieString)
	if err != nil {
		return nil, errors.New("invalid wookie string")
	}
	parts := strings.Split(string(decodedStr), ";")
	if len(parts) != 3 {
		return nil, errors.New("invalid wookie string")
	}
	id, err := strconv.ParseInt(parts[0], 0, 64)
	if err != nil {
		return nil, errors.New("invalid id in wookie string")
	}
	version, err := strconv.ParseInt(parts[1], 0, 64)
	if err != nil {
		return nil, errors.New("invalid version in wookie string")
	}
	microTs, err := strconv.ParseInt(parts[2], 0, 64)
	if err != nil {
		return nil, errors.New("invalid timestamp in wookie string")
	}
	timestamp := time.UnixMicro(microTs)

	return &Token{
		ID:        id,
		Version:   version,
		Timestamp: timestamp,
	}, nil
}
