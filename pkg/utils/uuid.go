package utils

import "github.com/google/uuid"

type UUID struct {
	value uuid.UUID
}

func GenerateUUID() UUID {
	return UUID{
		value: uuid.New(),
	}
}
