package jwtoken

import (
	"strconv"
	"time"

	"github.com/cristalhq/jwt/v3"
)

// Поля хранимые в токене
type UserClaims struct {
	jwt.RegisteredClaims
	Module    string `json:"module"`
	SessionID int    `json:"sId"`
	UserLogin string `json:"login"`
}

func (u UserClaims) GetUserID() int {
	uId, _ := strconv.Atoi(u.Subject)
	return uId
}

func (u UserClaims) GetLogin() string {
	return u.UserLogin
}

func (u UserClaims) GetSessionID() int {
	return u.SessionID
}

func (u UserClaims) Valid() bool {
	return u.ExpiresAt.After(time.Now())
}
