package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) *Service {
	return &Service{secret: []byte(secret), ttl: ttl}
}

func (s *Service) HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func (s *Service) CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func (s *Service) IssueToken(userID int64, username string, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"usr": username,
		"adm": isAdmin,
		"exp": time.Now().Add(s.ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
