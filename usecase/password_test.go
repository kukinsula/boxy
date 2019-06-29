package usecase

import (
	"testing"
)

func TestPassworder(t *testing.T) {
	passworder := NewPassworder(10)
	password := "Azerty1234."

	b, err := passworder.Hash(password)
	if err != nil {
		t.Errorf("Passworder.Hash should not fail %s", err)
		t.FailNow()
	}

	err = passworder.Compare(password, b)
	if err != nil {
		t.Errorf("Passworder.Compare should not fail %s", err)
		t.FailNow()
	}
}
