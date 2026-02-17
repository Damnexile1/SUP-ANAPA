package http

import "github.com/google/uuid"

func parseUUID(v string) uuid.UUID {
	id, err := uuid.Parse(v)
	if err != nil {
		return uuid.Nil
	}
	return id
}
