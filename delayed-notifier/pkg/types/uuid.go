package types

import (
	"fmt"

	"github.com/google/uuid"
)

type UUID struct {
	value uuid.UUID
}

func GenerateUUID() UUID {
	return UUID{
		value: uuid.New(),
	}

}


func NewUUID(id string) (UUID, error){
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid uuid '%s': %w", id, err)
	}
	return UUID{
		value: idUUID,
	}, nil
}

func(v * UUID) String() string {
	return v.value.String()
}


