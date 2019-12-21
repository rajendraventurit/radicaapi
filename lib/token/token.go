package token

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const defConfPath = "/etc/radica/jwt.json"

// Claims is a jwt claims struct
type Claims struct {
	UserID int64 `json:"user_id,omitempty"`
	jwt.StandardClaims
}

var localConf *config

type config struct {
	Secret      string `json:"secret"`
	key         []byte
	ExpireHours int `json:"expire_hours"`
	hours       time.Duration
}

// Configure will configure the logger using the defConfPath
// logger will attempt to open and write to the log file on each call
func Configure() error {
	f, err := os.Open(defConfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	conf := config{}
	err = json.NewDecoder(f).Decode(&conf)
	if err != nil {
		return err
	}
	if conf.ExpireHours < 1 {
		return fmt.Errorf("expire_hours must be greater than zero")
	}
	conf.hours = time.Duration(conf.ExpireHours) * time.Hour
	conf.key = []byte(conf.Secret)
	localConf = &conf

	return nil
}

// New returns a new token
func New(userid int64) (string, error) {
	if localConf == nil {
		if err := Configure(); err != nil {
			return "", err
		}
	}
	claims := Claims{UserID: userid}
	claims.Issuer = "apiserver"
	claims.ExpiresAt = time.Now().Add(localConf.hours).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(localConf.key)
}

// AuthToken extracts and returns the JWT token from the auth header
func AuthToken(r *http.Request) (*Claims, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return nil, fmt.Errorf("Authorization header format must be Bearer {token}")
	}
	h = strings.Trim(h, " \t\n\r")
	parts := strings.Split(h, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("Authorization header format must be Bearer {token}")
	}
	return Decode(parts[1])
}

// Decode will return a claim from a token string
func Decode(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return localConf.key, nil
		})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("Failed to decode token")
}

// SystemToken will generate a token for the system user. It is intended
// for use with external processes
func SystemToken(expire time.Time, secret []byte) (string, error) {
	claims := Claims{UserID: 1} // System user
	claims.Issuer = "apiserver"
	claims.Subject = "System API Token"
	// Slightly in the past (25s) to insure server timings don't invalidate token
	claims.IssuedAt = time.Now().Add(-25 * time.Second).Unix()
	claims.ExpiresAt = expire.Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
