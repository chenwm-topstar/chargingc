package oauth2

import (
	"fmt"

	googleidtokenverifier "github.com/movsb/google-idtoken-verifier"
)

func GoogleVerifyIdentityToken(ClientId string, token string) (*googleidtokenverifier.ClaimSet, error) {
	claims, err := googleidtokenverifier.Verify(token, ClientId)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Iss:\t%s\nSub:\t%s\nEmail:\t%s\nName:\t%s\nDomain:\t%s\n",
		claims.Iss, claims.Sub, claims.Email, claims.Name, claims.Domain)

	return claims, nil
}
