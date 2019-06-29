package entity

import (
	"github.com/satori/go.uuid"
)

func NewUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func FromStringToBytes(id string) ([]byte, error) {
	u, err := uuid.FromString(id)
	if err != nil {
		return []byte{}, nil
	}

	return u.Bytes(), nil
}

func FromBytes(data []byte) (string, error) {
	id, err := uuid.FromBytes(data)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
