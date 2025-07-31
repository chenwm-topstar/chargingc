package oauth2

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/chenwm-topstar/chargingc/utils/requests"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

var (
	GetApplePublicKeys = "https://appleid.apple.com/auth/keys"
	AppleUrl           = "https://appleid.apple.com"

	jKeys map[string][]JwtKeys
)

type (
	JwtClaims struct {
		CHash          string `json:"c_hash"`
		Email          string `json:"email"`
		EmailVerified  string `json:"email_verified"`
		AuthTime       int    `json:"auth_time"`
		NonceSupported bool   `json:"nonce_supported"`
		jwt.StandardClaims
	}

	JwtHeader struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}

	JwtKeys struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		Use string `json:"use"`
		Alg string `json:"alg"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
)

func AppleVerifyIdentityToken(ClientId, cliToken string, cliUserID string) (*JwtClaims, error) {
	cliTokenArr := strings.Split(cliToken, ".")
	if len(cliTokenArr) < 3 {
		return nil, errors.New("cliToken Split err")
	}

	cliHeader, err := jwt.DecodeSegment(cliTokenArr[0])
	if err != nil {
		return nil, err
	}

	var jHeader JwtHeader
	err = json.Unmarshal(cliHeader, &jHeader)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(cliToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return GetRSAPublicKey(jHeader.Kid)
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		if claims.StandardClaims.Issuer != AppleUrl || claims.StandardClaims.Audience != ClientId || claims.StandardClaims.Subject != cliUserID {
			return nil, errors.New("verify token info fail, info is not match")
		}

		return claims, nil
	}

	return nil, errors.New("token claims parse fail")
}

func GetRSAPublicKey(kid string) (*rsa.PublicKey, error) {
	if jKeys == nil || len(jKeys) <= 0 {
		jKeys = make(map[string][]JwtKeys)

		if err := requests.GetStruct(context.Background(), GetApplePublicKeys, &jKeys); err != nil {
			return nil, err
		}
	}

	var pubKey rsa.PublicKey
	for _, data := range jKeys {
		for _, val := range data {
			if val.Kid == kid {
				nByte, _ := base64.RawURLEncoding.DecodeString(val.N)
				nData := new(big.Int).SetBytes(nByte)

				eByte, _ := base64.RawURLEncoding.DecodeString(val.E)
				eData := new(big.Int).SetBytes(eByte)

				pubKey.N = nData
				pubKey.E = int(eData.Uint64())
				break
			}
		}
	}

	if pubKey.E <= 0 {
		return nil, errors.New("pubKey.E is nil")
	}

	return &pubKey, nil
}
