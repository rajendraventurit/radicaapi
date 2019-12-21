package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/rajendraventurit/radicaapi/lib/db"
)

func genURLToken(val string, key []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(val))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

/*
func verifyURLToken(val, tok string, key []byte) (bool, error) {
	token, err := base64.URLEncoding.DecodeString(tok)
	if err != nil {
		return false, err
	}
	exp := genURLToken(val, key)
	expected, err := base64.URLEncoding.DecodeString(exp)
	if err != nil {
		return false, err
	}
	return hmac.Equal(token, expected), nil
}
*/

// genResetToken will return a reset token
func genResetToken(qr db.Queryer, email string) ([]byte, error) {
	user, err := GetUserWithEmail(qr, email)
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("%v:%v", user.UserID, email)
	mac := hmac.New(sha256.New, []byte(resetKey))
	_, err = mac.Write([]byte(msg))
	return mac.Sum(nil), err
}

// encodeResetToken returns a url encoded version of the reset token
func encodeResetToken(tok []byte) string {
	return base64.URLEncoding.EncodeToString(tok)
}

// verifyResetToken will verify a reset token
func verifyResetToken(qr db.Queryer, email, tok string) (bool, error) {
	tokHMAC, err := base64.URLEncoding.DecodeString(tok)
	if err != nil {
		return false, err
	}
	expectedHMAC, err := genResetToken(qr, email)
	if err != nil {
		return false, err
	}
	return hmac.Equal(tokHMAC, expectedHMAC), nil
}
