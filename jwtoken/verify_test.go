package jwtoken

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/cristalhq/jwt/v3"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Время токена еще не прошло. Он еще валидный
func TestVerifyOK(t *testing.T) {
	key, _ := GenerateKey()

	expTime := time.Now().Add(time.Second * 40)
	token, err := Create(key, 4, "user52", 4, "test module", expTime)
	assert.Nil(t, err)
	assert.Equal(t, true, VerifyToken(&key.PublicKey, token))
}

// Время токена еще прошло он уже не валидный
func TestVerifyFailedExpTimeSpent(t *testing.T) {
	key, _ := GenerateKey()

	expTime := time.Now().Add(time.Second * 40 * -1)
	token, err := Create(key, 3, "user52", 5, "test module", expTime)
	assert.Nil(t, err)
	assert.Equal(t, false, VerifyToken(&key.PublicKey, token))
}

func TestRSAExample(t *testing.T) {
	// СОздаем ключи
	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)
	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)
	//------------------------
	// Создаем токен
	signer, _ := jwt.NewSignerRS(jwt.RS512, privateKey1)
	builder := jwt.NewBuilder(signer)

	var expDate jwt.NumericDate
	expDate.Time = time.Date(2021, 10, 10, 11, 11, 10, 0, time.Local)

	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"some"},
			ExpiresAt: &expDate,
		},
		Module: "sss",
	}
	token, err := builder.Build(claims)
	assert.Nil(t, err)
	{
		// Расшифровываем токен и проверяем с помощью Private ключа
		verifier, err := jwt.NewVerifierRS(jwt.RS512, &privateKey1.PublicKey)
		assert.Nil(t, err)
		// Разбираем токен
		nekot, err := jwt.ParseString(token.String())
		assert.Nil(t, err)
		// Проверяем подпись
		err = verifier.Verify(nekot.Payload(), nekot.Signature())
		assert.Nil(t, err)
	}
	// Попробуем использовать другой ключ
	{
		// Расшифровываем токен и проверяем с помощью Private ключа
		verifier, err := jwt.NewVerifierRS(jwt.RS512, &privateKey2.PublicKey)
		assert.Nil(t, err)
		// Разбираем токен
		nekot, err := jwt.ParseString(token.String())
		assert.Nil(t, err)
		// Проверяем подпись - ошибка
		err = verifier.Verify(nekot.Payload(), nekot.Signature())
		assert.NotNil(t, err)
	}
}
