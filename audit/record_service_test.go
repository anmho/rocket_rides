package audit_test

import (
	"context"
	audit "github.com/anmho/idempotent-rides/audit"
	"github.com/anmho/idempotent-rides/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	ExistingTestRecord = &audit.Record{
		ID:       4321,
		Action:   "created",
		Data:     []byte("{}"),
		OriginIP: "127.0.0.1/32",
		Resource: audit.Resource{
			ID:   1441,
			Type: "ride",
		},
		UserID: 123,
	}
	NewTestRecord = &audit.Record{
		ID:        1,
		Action:    "ride_charged",
		CreatedAt: time.Time{},
		Data:      []byte("{\"target\": {\"\": {\"lat\": 10, \"long\": 20}}}"),
		OriginIP:  "127.0.0.1/32",
		Resource: audit.Resource{
			ID:   1441,
			Type: "ride",
		},
		UserID: 123,
	}
)

func assertEqualRecord(t *testing.T, expected, record *audit.Record) {
	assert.Equal(t, expected.ID, record.ID)
	assert.Equal(t, expected.Action, record.Action)
	assert.Equal(t, expected.Data, record.Data)
	assert.Equal(t, expected.OriginIP, record.OriginIP)
	assert.Equal(t, expected.Resource, record.Resource)
	assert.Equal(t, expected.UserID, record.UserID)
}

func TestService_GetRecord(t *testing.T) {
	tests := []struct {
		desc     string
		recordID int

		expectedErr    bool
		expectedRecord *audit.Record
	}{
		{
			desc:     "happy path: get existing record by id",
			recordID: ExistingTestRecord.ID,

			expectedRecord: ExistingTestRecord,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)
			ctx := context.Background()
			tx := test.MakeTx(t, ctx, db)
			s := audit.MakeService()

			record, err := s.GetRecord(ctx, tx, tc.recordID)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
				assertEqualRecord(t, tc.expectedRecord, record)
			}
		})
	}

}
func TestService_CreateRecord(t *testing.T) {
	tests := []struct {
		desc   string
		record *audit.Record

		expectedErr    bool
		expectedRecord *audit.Record
	}{
		{
			desc:   "happy path: create record that doesn't exist yet",
			record: NewTestRecord,

			expectedErr:    false,
			expectedRecord: NewTestRecord,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)
			ctx := context.Background()
			tx := test.MakeTx(t, ctx, db)
			s := audit.MakeService()
			record, err := s.CreateRecord(ctx, tx, tc.record)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
				assertEqualRecord(t, tc.expectedRecord, record)
			}
		})
	}
}

//func TestService_UpdateRecord(t *testing.T) {
//	tests := []struct {
//		desc   string
//		record *audit.Record
//
//		expectedErr    bool
//		expectedRecord *audit.Record
//	}{
//		{
//			desc: "happy path: update existing record",
//			record: &audit.Record{
//				ID:        ExistingTestRecord.ID,
//				Action:    ExistingTestRecord.Action,
//				CreatedAt: ExistingTestRecord.CreatedAt,
//				Data:      ExistingTestRecord.Data,
//				OriginIP:  ExistingTestRecord.OriginIP,
//				Resource:  ExistingTestRecord.Resource,
//				UserID:    ExistingTestRecord.UserID,
//			},
//		},
//		{
//			desc: "error case: update non-existent record",
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.desc, func(t *testing.T) {
//			db := test.MakePostgres(t)
//			ctx := context.Background()
//			tx := test.MakeTx(t, ctx, db)
//			s := MakeService()
//
//			record, err := s.UpdateRecord(ctx, tx, tc.record)
//			if tc.expectedErr {
//				assert.Error(t, err)
//				assert.Nil(t, record)
//			} else {
//				assert.NoError(t, err)
//				assert.NotNil(t, record)
//				assertEqualRecord(t, tc.expectedRecord, record)
//			}
//		})
//	}
//}
//func TestService_DeleteRecord(t *testing.T) {
//	tests := []struct {
//		desc     string
//		recordID int
//
//		expectedAffectedRow bool
//		expectedErr         bool
//	}{
//		{
//			desc:                "happy path: delete existing record",
//			recordID:            ExistingTestRecord.ID,
//			expectedAffectedRow: true,
//		},
//		{
//			desc:                "error path: delete non-existent record",
//			recordID:            123123,
//			expectedErr:         false,
//			expectedAffectedRow: false,
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.desc, func(t *testing.T) {
//			db := test.MakePostgres(t)
//			ctx := context.Background()
//			tx := test.MakeTx(t, ctx, db)
//			s := MakeService()
//			ok, err := s.DeleteRecord(ctx, tx, 1)
//			if tc.expectedErr {
//				assert.Error(t, err)
//				assert.False(t, ok)
//			} else {
//				assert.NoError(t, err)
//				assert.Equal(t, ok, tc.expectedAffectedRow)
//			}
//		})
//	}
//}
