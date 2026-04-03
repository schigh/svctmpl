package convert

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// StringToUUID converts a string UUID to pgtype.UUID.
func StringToUUID(s string) pgtype.UUID {
	var id pgtype.UUID
	_ = id.Scan(s)
	return id
}

// UUIDToString converts a pgtype.UUID to a standard UUID string.
func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
