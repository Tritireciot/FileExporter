package jwtoken

import (
	"bytes"
	"encoding/pem"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Проверка функции разбора токена в значении заголовка Authorization
func TestParseHeaderToken(t *testing.T) {
	// Корректно вырезаем
	assert.Equal(t, "token_test123_sahgjfa", ParseHeaderToken(authHeaderKeyString+"token_test123_sahgjfa"))
	// Не смогли найт токен из-за пробела
	assert.Equal(t, "  token_test123_sahgjfa", ParseHeaderToken(authHeaderKeyString+"  token_test123_sahgjfa"))
	// Пустая строка
	assert.Equal(t, "", ParseHeaderToken(""))
	// Отсутствует токен
	assert.Equal(t, "", ParseHeaderToken(authHeaderKeyString))
	//мусор
	assert.Equal(t, `VPKW6PhjepO1pthrnoxxp2kD3p2RQ8yLNZqnnVcfkm7oSZwBsUFa1LpC7MJ1wz7bsYXXyUGJQUg0DpTzjdNhKkTexXUdGOS9bYK1ubpt16efJXwMgcWo2jWnRawgdmRL3mudrejuSSRh
zWXIdM4mZCnURBsaAp8xM14I0gQBGRbr2THQYVSKOb4GIghZoAc80xRSMd VUyAqLuPmxOqFUHHLaEhSAmJrmNfMpFDgF4jJLaU1Wa4GlQTbcr85nKWeDSY uzpfnQdwpgEsbMNvxwAs5DbCBbkFV8TbnE66ODxOwt7XHhBFRtixYm0f1QqigU
enkHQ75CylGZAPlQ5dLEh0S7gqEC9YsJd5q3reLHv2m5QaHu2xTqtyggtG6plm1b3n5K8CHYeSAjmuVXEvKYpTNUkXZARv5VKBoHXC1vNSafwnBqXXZA7Abw9VkcyixoiEvdTCRjucKwgkhj9oNPnLZrDO805vZbV8xViCrE9hnClRP7VMF6`,
		ParseHeaderToken(`VPKW6PhjepO1pthrnoxxp2kD3p2RQ8yLNZqnnVcfkm7oSZwBsUFa1LpC7MJ1wz7bsYXXyUGJQUg0DpTzjdNhKkTexXUdGOS9bYK1ubpt16efJXwMgcWo2jWnRawgdmRL3mudrejuSSRh
zWXIdM4mZCnURBsaAp8xM14I0gQBGRbr2THQYVSKOb4GIghZoAc80xRSMd VUyAqLuPmxOqFUHHLaEhSAmJrmNfMpFDgF4jJLaU1Wa4GlQTbcr85nKWeDSY uzpfnQdwpgEsbMNvxwAs5DbCBbkFV8TbnE66ODxOwt7XHhBFRtixYm0f1QqigU
enkHQ75CylGZAPlQ5dLEh0S7gqEC9YsJd5q3reLHv2m5QaHu2xTqtyggtG6plm1b3n5K8CHYeSAjmuVXEvKYpTNUkXZARv5VKBoHXC1vNSafwnBqXXZA7Abw9VkcyixoiEvdTCRjucKwgkhj9oNPnLZrDO805vZbV8xViCrE9hnClRP7VMF6`))
	// Отсутствует токен - неправильный ключ
	assert.Equal(t, "Bearer", ParseHeaderToken("Bearer"))
	// короткий токен
	assert.Equal(t, "1", ParseHeaderToken(authHeaderKeyString+"1"))
	// длинный токен
	assert.Equal(t, "fhksadfhsg827f4398h21f/2.v34812v34g182g634871628dh424f.214yv12948721fg7834y2837r8723jgjsdfg018fgshdgafjkgsadfgsbdfgsa0ft18ft183tfbgasfbgsf108tf13fbd",
		ParseHeaderToken(authHeaderKeyString+"fhksadfhsg827f4398h21f/2.v34812v34g182g634871628dh424f.214yv12948721fg7834y2837r8723jgjsdfg018fgshdgafjkgsadfgsbdfgsa0ft18ft183tfbgasfbgsf108tf13fbd"))

	token := ParseHeaderToken(`NewsServer=VcWsYbC1jdpVeZP2a7C1xe7ObPekrQIhLTpBEzrRQW0=; Authorization="Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzg0OTA1ODY1LCJtb2R1bGUiOiIiLCJzSWQiOjE1MzI3MywibG9naW4iOiJhZG1pbiJ9.EwTahNm0hPE2l4GBYxOlUikBIQTGuPuFSfKNNnLNL2q9vvk-0hhBPdJq0_sclWITgCPf-dqtE7gEbKne20cuywVxWw9J-tvToDYEhS1BtMdSzs1LcxYF4VR81KPoLGfESvJ8Mk_tTnALc2-TWtXKCIrn4m0Hz3aNNMOycOI1NP-UjhUHONF-jSiGelsDn--jW5vEXPZ_K_QJhTyqO98zcEypN_kRSmOP3xtCXR3HtAX1OX8RrzmhLf43-dV4vdkksVLP_tGKelFXDpAiZejEIYW2lQKbZveLIuQ6UIaYiLt7RZJbRzFL1__ZRiWWb2PIsj6iiGOKQcGB8z4I3bgS5A"`)
	assert.Equal(t, `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzg0OTA1ODY1LCJtb2R1bGUiOiIiLCJzSWQiOjE1MzI3MywibG9naW4iOiJhZG1pbiJ9.EwTahNm0hPE2l4GBYxOlUikBIQTGuPuFSfKNNnLNL2q9vvk-0hhBPdJq0_sclWITgCPf-dqtE7gEbKne20cuywVxWw9J-tvToDYEhS1BtMdSzs1LcxYF4VR81KPoLGfESvJ8Mk_tTnALc2-TWtXKCIrn4m0Hz3aNNMOycOI1NP-UjhUHONF-jSiGelsDn--jW5vEXPZ_K_QJhTyqO98zcEypN_kRSmOP3xtCXR3HtAX1OX8RrzmhLf43-dV4vdkksVLP_tGKelFXDpAiZejEIYW2lQKbZveLIuQ6UIaYiLt7RZJbRzFL1__ZRiWWb2PIsj6iiGOKQcGB8z4I3bgS5A`, token)

}

// Проверка правильности сохранения информации в токене
func TestGetUserClaims(t *testing.T) {
	userName := "user123test"
	userId := 123
	sessionId := 543
	moduleName := "module name test"
	expTime := time.Now().Add(time.Second * 50)
	key, _ := GenerateKey()

	token, err := Create(key, userId, userName, sessionId, moduleName, expTime)
	assert.Nil(t, err)

	claims, err := GetUserClaims(&key.PublicKey, token)
	assert.Nil(t, err)
	assert.Equal(t, userName, claims.GetLogin())
	assert.Equal(t, userId, claims.GetUserID())
	assert.Equal(t, sessionId, claims.GetSessionID())
	assert.Equal(t, moduleName, claims.Module)
	assert.Equal(t, expTime.Unix(), claims.ExpiresAt.Time.Unix())
}

// Проверка разбора токена из готового сообщения
func TestParseFromBeginning(t *testing.T) {
	// Формируем токен
	userName := "user123test"
	moduleName := "module name test"
	expTime := time.Now().Add(time.Second * 50)
	key, _ := GenerateKey()

	token, err := Create(key, 3, userName, 2, moduleName, expTime)
	assert.Nil(t, err)

	buf := []byte("sdfasdfs")
	req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(buf))
	assert.Equal(t, nil, err)

	req.Header.Set("Authorization", "Bearer "+token)

	assert.Equal(t, true, ParseHeaderBearerAndVerify(&key.PublicKey, req.Header.Get("Authorization")))
}

func TestCreateToken(t *testing.T) {
	// Формируем токен

	pm, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2P2g+OEbpF+sDAohlucx7tGYJiEOo2mD6iwX4Y84AYtjX/QI
h6xiEmdfw4vRPCLAXx69K66AS931/pcOiLM+nppECQcrKqAmhOu4Ip9VfmSBeJ6m
eYVt9J+f+tuHW0pm4G2q7AhoyBfkyElahYqpVsI+5IePK9qJhhml0s6WX5wmcwL0
Wn3Bzk/4a3syHDsFB7nvGh6jyMKGzyOxS02oyLqYhSgq3Zyx4Tn8vwGSFMX4KcEC
0si4wWGXTNakwWYxWYBAFcRTw7rfG64b8ITHHiSBjMeRBlVSoRJ7bs6bxY/AFV58
tKfz2radCgFDSZfA5v4axvm59UUHgBmJcHwypQIDAQABAoIBAEKi8PI9Px3le5Je
8h3DdiQfHZhoAnTQjIA3dkYAk3R199iZupzfpWZ9dH06zNCo42bSq8lkV2X1DfxX
K1FzqkFOoqIbH3iBohKjyPJo9/pOpywBnKIpBbFf6+M/03uHh7xYMWs20ebQ36Na
U4A7KvHHyUSpFKClBiK2caQhaTCG1J22rRxsEvzTm8iPIiTaiCabhSigp/gSfx8C
N4MDvmzr45dLpBuTtJzcp+UmgR7TERnL+219brjQiMQDvsa5Rk5nfO0AQ5RqTDsr
4z0YY+BqDPkXK0wpdivmL6O1H7JajxleeYPiuNvlXJpZooY2NSjrg88g5meVh1aF
28UGiykCgYEA4AenmUOgLKNIE/OdZGw3X0W2A4bIwx1v02xhRu2TMISfGL8sTv/P
f37lE+TWh1LjkeUyIKAafFPG6pM+4uyKxWQsVFI0IeItd7d5vQpbCdBVMOaFXw1o
j07rz6RhRu4vn2f3eKGdS8ZXqZJBbp4/VrePxXJpJl8c+Kwbvc2iQScCgYEA9/TR
FJT0iW3NcxtOFeK7g1lPQA8aIzXWIfQA1oUrYODjculHS5UL2jptEVRMl/TWeB2S
GU6gwBNafuAnJp4it1ce6b96Ty3SfW3RvFOiilaC2dtZd7YPQUf4enbmE3FB1fy6
mYC1k55AXbuyIg8V0lI5VUVxzOVm7uaME338NVMCgYBgoIyWW4e7mRWenWXWiaJ8
cknmTX1MQucXrthqSlBBDgK9Hr/Stx1dZXMS2JH7PjIvnEa5sbSayVuzk5z9LX9R
UjqYh/g6YR6xUy6r7cqEeho0hEkkTVk67pRhNApNGLDrtWEU48g/haYL2qxkNNcm
5Pea9xUJWt7ZFwrEG+yO4wKBgQCvyXKJAukynROwbUU4otuJTUGwCoTfPYWn4JLP
gu2z6vuVNekDnpEej5lPVdJPUJbT5TL0mhfA1Hetx0A6UpYOIMebs9IEXFoD7l1p
BGoRZS+zP5z4D6xU/a8dMzn9wqeIC3pG5UbVdrXvPegV2VgBGaXn3CnHt0L4T54z
G3g4XQKBgGMehSDWE1A9Nvk2ckabVgDhbHzmT1I3XquiAhWcdQZRp4oXXwlBexFv
jlMUJjsfnrHiq8YPTyywj/ANEIUNi1gbF+lhxQReUEEfiMVZQgO/sT0h4mUWZGDV
pJl89oMFAWoZ2It/ogauTnOJguqINugBK/QD9iV2RA03Oc3KLAAz
-----END RSA PRIVATE KEY-----`))
	pkey, err := LoadPrivateEM(pm)
	assert.Nil(t, err)

	accessToken, err := Create(pkey, 1, "admin", 1, "Test", time.Now().Add(time.Hour*24*365*12))
	assert.Nil(t, err)
	log.Print(accessToken)
}
