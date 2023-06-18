package wookie

import "github.com/pkg/errors"

type Token interface {
	Deserialize() Wookie
	Serialize() string
}

type TokenKey struct{}

type BasicToken struct {
	// OSS version
}

// Get wookie token from incoming http request (header)
// func FromRequest(r *http.Request) BasicToken {
// 	return BasicToken{}
// }

// func Create() {

// }

// func Validate() {

// }

func Deserialize(wookieString string) (BasicToken, error) {
	if wookieString == "" {
		return BasicToken{}, errors.New("wookie string empty")
	}
	return BasicToken{}, nil
}
