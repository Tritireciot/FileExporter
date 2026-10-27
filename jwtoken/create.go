package jwtoken

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"time"

	"github.com/cristalhq/jwt/v3"
)

const (
	DefAccessExpirationDurationMin  = 10 * time.Minute      // Длительность валидность access токена в минутах (по умолчанию)
	DefRefreshExpirationDurationMin = 60 * 24 * time.Minute // Длительность валидность refresh токена в минутах (по умолчанию)

	AuthorizationHeaderTag = "Authorization"
)

var (
	accessExpirationDurationMin  = DefAccessExpirationDurationMin
	refreshExpirationDurationMin = DefRefreshExpirationDurationMin
)

func SetAccessExpirationPeriod(p time.Duration) {
	accessExpirationDurationMin = p
}

func SetRefreshExpirationPeriod(p time.Duration) {
	refreshExpirationDurationMin = p
}

func GetAccessExpirationPeriod() time.Duration {
	return accessExpirationDurationMin
}

func GetRefreshExpirationPeriod() time.Duration {
	return refreshExpirationDurationMin
}

// ---------------------------------------------------------------------------
// Функция формирования времени истечения access токена
func GetAcessExpDate() time.Time {
	return GetExpDate(time.Now(), accessExpirationDurationMin)
}

// ---------------------------------------------------------------------------
// Функция формирования времени истечения refresh токена
func GetRefreshExpDate() time.Time {
	return GetExpDate(time.Now(), refreshExpirationDurationMin)
}

// ---------------------------------------------------------------------------
// Сформируем время валидности токена
func GetExpDate(nowTime time.Time, expDurationMin time.Duration) time.Time {
	return nowTime.Add(expDurationMin)
}

// ---------------------------------------------------------------------------
// Формирование токена
func Create(key *rsa.PrivateKey, userID int, userLogin string, sessionID int, module string, exp time.Time) (string, error) {
	signer, err := jwt.NewSignerRS(jwt.RS512, key)
	if err != nil {
		return "", err
	}
	builder := jwt.NewBuilder(signer)

	var expDate jwt.NumericDate
	expDate.Time = exp

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &expDate,
			Subject:   fmt.Sprintf("%d", userID),
		},
		Module:    module,
		SessionID: sessionID,
		UserLogin: userLogin,
	}
	token, err := builder.Build(claims)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return token.String(), nil
}

// ---------------------------------------------------------------------------
// Создадим ключ
func GenerateKey() (key *rsa.PrivateKey, err error) {
	key, err = rsa.GenerateKey(rand.Reader, 2048)

	return key, err
}
