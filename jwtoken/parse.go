package jwtoken

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cristalhq/jwt/v3"
)

var (
	authHeaderKeyString = "Bearer "
)

// Распарсить токен и получить хранимую информацию
func GetUserClaims(key *rsa.PublicKey, tokenData string) (UserClaims, error) {
	var claims UserClaims

	verifier, err := jwt.NewVerifierRS(jwt.RS512, key)
	if err != nil {

		return claims, fmt.Errorf("can't create verifyer hs")
	}
	// parse a Token
	token, err := jwt.ParseString(tokenData)
	if err != nil {
		return claims, fmt.Errorf("can't parse token")
	}

	err = verifier.Verify(token.Payload(), token.Signature())
	if err != nil {
		return claims, fmt.Errorf("can't verify token")
	}

	errClaims := json.Unmarshal(token.RawClaims(), &claims)
	if errClaims != nil {
		return claims, fmt.Errorf("can't unmarshal user claims")
	}

	return claims, nil
}

// Разбор токена из заголовка "Bearer <token>"
func ParseHeaderToken(tokenHeader string) string {
	idx := strings.Index(tokenHeader, authHeaderKeyString)
	if idx >= 0 {
		idx = idx + len(authHeaderKeyString)

		token := tokenHeader[idx:]

		idx := strings.Index(token, `"`)
		if idx >= 0 {
			return token[:idx]
		}
		return token
	}
	return tokenHeader
}

// Разобрать заголовок и проверить токен
func ParseHeaderBearerAndVerify(key *rsa.PublicKey, tokenHeader string) bool {
	return VerifyToken(key, ParseHeaderToken(tokenHeader))
}
