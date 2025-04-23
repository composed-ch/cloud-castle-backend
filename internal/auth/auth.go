package auth

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "utmostsecret"

var (
	signingMethod *jwt.SigningMethodHMAC = jwt.SigningMethodHS512
	authHeader    *regexp.Regexp         = regexp.MustCompilePOSIX("^Bearer (.+)$")
)

type Handler func(http.ResponseWriter, *http.Request)

func IssueToken(username string) (string, error) {
	iat := time.Now()
	exp := iat.Add(time.Hour * 24)
	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"sub": username,
		"iat": iat.Unix(),
		"exp": exp.Unix(),
	})
	return token.SignedString([]byte(secret))

}

func Authenticated(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		matches := authHeader.FindStringSubmatch(authorization)
		if len(matches) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tokenStr := matches[1]
		_, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{signingMethod.Alg()}))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
