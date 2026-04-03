package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12
// Hash passs func 
// cost=12 — slow enough to prevent brute force.
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil{
		return " ", fmt.Errorf("hash password : %w", err)
	}
	return string(bytes), nil
}

// CheckPassword func
// return = nil if match, err if wrong 
func CheckPassword(plain, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}