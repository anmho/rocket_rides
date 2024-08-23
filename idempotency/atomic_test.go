package idempotency

import (
	"testing"
)

func Test_AtomicPhase(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{
			desc: "error path: nil idempotency key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			//db, cleanup := testfixtures.MakePostgres(t)
			//t.Cleanup(func() { cleanup() })
			//key, err := AtomicPhase(ctx, nil, db)
			//require.NoError(t)
		})
	}
}
