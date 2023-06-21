package wookie

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ClientToken context key
type ClientTokenKey struct{}

type Token struct {
	ID        int64
	Version   int64
	Timestamp time.Time
}

// Get string representation of token (to set as header)
func (t Token) String() string {
	s := fmt.Sprintf("%d;%d;%d", t.ID, t.Version, t.Timestamp.UnixMicro())
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// De-serialize token from string (from header)
func FromString(wookieString string) (Token, error) {
	if wookieString == "" {
		return Token{}, errors.New("empty wookie string")
	}
	decodedStr, err := base64.StdEncoding.DecodeString(wookieString)
	if err != nil {
		return Token{}, errors.New("invalid wookie string")
	}
	parts := strings.Split(string(decodedStr), ";")
	if len(parts) != 3 {
		return Token{}, errors.New("invalid wookie string")
	}
	id, err := strconv.ParseInt(parts[0], 0, 64)
	if err != nil {
		return Token{}, errors.New("invalid id in wookie string")
	}
	version, err := strconv.ParseInt(parts[1], 0, 64)
	if err != nil {
		return Token{}, errors.New("invalid version in wookie string")
	}
	microTs, err := strconv.ParseInt(parts[2], 0, 64)
	if err != nil {
		return Token{}, errors.New("invalid timestamp in wookie string")
	}
	timestamp := time.UnixMicro(microTs)

	return Token{
		ID:        id,
		Version:   version,
		Timestamp: timestamp,
	}, nil
}
