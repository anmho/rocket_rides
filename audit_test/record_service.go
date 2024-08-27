package audit

import (
	"context"
	"database/sql"
)

type Service interface {
	GetRecord(ctx context.Context, tx *sql.Tx, recordID int) (*Record, error)
	CreateRecord(ctx context.Context, tx *sql.Tx, record *Record) (*Record, error)
	//UpdateRecord(ctx context.Context, tx *sql.Tx, record *Record) (*Record, error)
	//DeleteRecord(ctx context.Context, tx *sql.Tx, id int) (bool, error)
}

func MakeService() Service {
	return &service{}
}

var _ Service = (*service)(nil)

type service struct {
}

func (s *service) GetRecord(ctx context.Context, tx *sql.Tx, id int) (*Record, error) {
	query := `
	SELECT 
	    id, created_at,
	    action, data, origin_ip, 
	    resource_id, resource_type, user_id
	FROM rocket_rides.public.audit_records
	WHERE id = $1
	;
	`

	row := tx.QueryRowContext(ctx, query, id)

	var record Record
	err := row.Scan(
		&record.ID, &record.CreatedAt,
		&record.Action, &record.Data, &record.OriginIP,
		&record.Resource.ID, &record.Resource.Type, &record.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// CreateRecord creates a new record with the given properties.
func (s *service) CreateRecord(ctx context.Context, tx *sql.Tx, record *Record) (*Record, error) {
	query := `
	INSERT INTO rocket_rides.public.audit_records (
		action, data, origin_ip, 
		resource_id, resource_type, 
		user_id
	) VALUES (
		$1, $2, $3,
	    $4, $5,
		$6
	) RETURNING 
	    id, created_at, origin_ip,  
		action, data, 
	    resource_id, resource_type, 
	    user_id
	;
	`

	row := tx.QueryRowContext(ctx, query,
		record.Action, record.Data, record.OriginIP,
		record.Resource.ID, record.Resource.Type,
		record.UserID,
	)

	var newRecord Record
	err := row.Scan(
		&newRecord.ID, &newRecord.CreatedAt, &newRecord.OriginIP,
		&newRecord.Action, &newRecord.Data,
		&newRecord.Resource.ID, &newRecord.Resource.Type,
		&newRecord.UserID,
	)

	if err != nil {
		return nil, err
	}

	return &newRecord, nil
}

//// UpdateRecord updates the fields record with the corresponding ID. CreatedAt is ignored.
//func (s *service) UpdateRecord(ctx context.Context, tx *sql.Tx, record *Record) (*Record, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//// DeleteRecord dele
//func (s *service) DeleteRecord(ctx context.Context, tx *sql.Tx, id int) (bool, error) {
//	//TODO implement me
//	panic("implement me")
//}
