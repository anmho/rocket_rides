package audit

import (
	"time"
)

type Resource struct {
	ID   int
	Type string
}
type Record struct {
	ID int
	// action taken, for example "created"
	Action    string
	CreatedAt time.Time
	Data      []byte
	OriginIP  string

	Resource Resource
	// UserID is the ID of the user that initiated this record
	UserID int
}

func NewRecord(action string, data []byte, OriginIP string, resource Resource, userID int) *Record {
	return &Record{
		Action:   action,
		Data:     data,
		OriginIP: OriginIP,
		Resource: resource,
		UserID:   userID,
	}
}
