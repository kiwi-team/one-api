package klingai

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

var klingaiTokens sync.Map
var expSeconds int64 = 60 * 30

func ParseConfig(config string) (secretId string, secretKey string, err error) {
	parts := strings.Split(config, "|")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid klingai config: %s", config)
		return
	}
	secretId = parts[0]
	secretKey = parts[1]
	return secretId, secretKey, nil
}

func GetToken(apiKey string) (string, error) {
	ak, sk, err := ParseConfig(apiKey)

	if err != nil {
		return "", err
	}

	expMillis := time.Now().Add(time.Duration(expSeconds)*time.Second).UnixNano() / 1e6
	expiryTime := time.Now().Add(time.Duration(expSeconds) * time.Second)

	payload := jwt.StandardClaims{
		Issuer:    ak,
		ExpiresAt: expMillis,
		NotBefore: time.Now().Add(time.Second * -5).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// 将 secret key 转换为字节切片
	secretKeyBytes := []byte(sk)

	tokenString, err := token.SignedString(secretKeyBytes)
	if err != nil {
		return "", err
	}

	klingaiTokens.Store(apiKey, tokenData{
		Token:      tokenString,
		ExpiryTime: expiryTime,
	})

	return tokenString, nil
}
