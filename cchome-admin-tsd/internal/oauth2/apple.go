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
		CHash string `json:"c_hash"`
		Email string `json:"email"`
		// EmailVerified  string `json:"email_verified"`
		AuthTime       int  `json:"auth_time"`
		NonceSupported bool `json:"nonce_supported"`
		jwt.StandardClaims
		// jwt中clamis的基础字段，上面几个为苹果官方自定义的字段，很多人不知
		// 道除基础字段以外的第三方自定义字段如何接受，只需要像上面一样在基础字段同
		// 级定义就行
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

// AppleVerifyIdentityToken 认证客户端传递过来的token是否有效
func AppleVerifyIdentityToken(ClientId, cliToken string, cliUserID string) (*JwtClaims, error) {
	// 数据由 头部、载荷、签名 三部分组成
	cliTokenArr := strings.Split(cliToken, ".")
	if len(cliTokenArr) < 3 {
		return nil, errors.New("cliToken Split err")
	}

	// 解析cliToken的header获取kid
	cliHeader, err := jwt.DecodeSegment(cliTokenArr[0])
	if err != nil {
		return nil, err
	}

	var jHeader JwtHeader
	err = json.Unmarshal(cliHeader, &jHeader)
	if err != nil {
		return nil, err
	}

	// 效验pubKey 及 token
	token, err := jwt.ParseWithClaims(cliToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return GetRSAPublicKey(jHeader.Kid)
	})

	if err != nil {
		return nil, err
	}

	// 信息验证
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		if claims.StandardClaims.Issuer != AppleUrl || claims.StandardClaims.Audience != ClientId || claims.StandardClaims.Subject != cliUserID {
			return nil, errors.New("verify token info fail, info is not match")
		}

		return claims, nil
	}

	return nil, errors.New("token claims parse fail")
}

/*
	GetRSAPublicKey 向苹果服务器获取解密signature所需要用的publicKey，苹果官方

返回的公钥不止一个，可能有多个，只需要像下面一样通过和identifyToken的header里的kid
比对，找匹配到的那一个使用就行。jwt总共分三段，前两段其实只需要通过base64直接反解就可
以获取到内容了，这个也是很多同学不知道的
*/
func GetRSAPublicKey(kid string) (*rsa.PublicKey, error) {
	if jKeys == nil || len(jKeys) <= 0 {
		jKeys = make(map[string][]JwtKeys)

		if err := requests.GetStruct(context.Background(), GetApplePublicKeys, &jKeys); err != nil {
			return nil, err
		}
	}

	// 获取验证所需的公钥
	var pubKey rsa.PublicKey
	// 通过cliHeader的kid比对获取n和e值 构造公钥
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
