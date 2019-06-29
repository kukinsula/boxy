package usecase

import (
	"golang.org/x/crypto/bcrypt"
)

type Passworder struct {
	Cost int
}

func NewPassworder(cost int) *Passworder {
	return &Passworder{
		Cost: cost,
	}
}

func (passworder *Passworder) Hash(password string) ([]byte, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), passworder.Cost)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

func (passworder *Passworder) Compare(encrypted, password []byte) error {
	return bcrypt.CompareHashAndPassword(encrypted, password)
}
