package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}

	parsed, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return ""
	}

	return parsed.String()
}

func stringToUUID(value string) (pgtype.UUID, error) {
	if value == "" {
		return pgtype.UUID{}, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

func timeFrom(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}

func textFrom(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
