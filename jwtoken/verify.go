package jwtoken

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"time"

	"github.com/cristalhq/jwt/v3"
)

// Проверка токена на валидность по целостности и по дате ExpiresAt
func VerifyToken(key *rsa.PublicKey, tokenData string) bool {
	verifier, err := jwt.NewVerifierRS(jwt.RS512, key)
	if err != nil {
		log.Print("Can't create verifyer hs!")
		return false
	}
	// parse a Token
	token, err := jwt.ParseString(tokenData)
	if err != nil {
		log.Print("Can't parse token!")
		return false
	}

	err = verifier.Verify(token.Payload(), token.Signature())
	if err != nil {
		log.Print("Can't verify token!")
		return false
	}

	// get standard claims
	var claims UserClaims
	errClaims := json.Unmarshal(token.RawClaims(), &claims)
	if errClaims != nil {
		log.Print("Can't unmarshal user claims!")
		return false
	}

	return claims.ExpiresAt.After(time.Now())
}
