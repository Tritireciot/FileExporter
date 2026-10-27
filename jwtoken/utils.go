package jwtoken

import (
	"crypto/rsa"
	"net/http"
)

func ParseRequestToken(key *rsa.PublicKey, r *http.Request) (UserClaims, error) {
	tokenHeader := ""
	if c, err := r.Cookie("Authorization"); err == nil {
		tokenHeader = c.Value
	} else {
		tokenHeader = r.Header.Get("Authorization")
	}
	ParseHeaderToken(tokenHeader)

	return GetUserClaims(key, tokenHeader)
}
