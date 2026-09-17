package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(v string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
	return string(b), e
}
func comparePassword(hash, value string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)) == nil
}
func generateOTP() (string, error) {
	var b [4]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}
func otpHash(challengeID, code, pepper string) string {
	h := sha256.Sum256([]byte(challengeID + ":" + code + ":" + pepper))
	return hex.EncodeToString(h[:])
}
func secureEqual(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
