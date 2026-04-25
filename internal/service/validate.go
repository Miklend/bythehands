package service

import "github.com/google/uuid"

func validateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
