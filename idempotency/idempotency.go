package idempotency

import "net/http"

const (
	HeaderKey = "Idempotency-Key"
)

type RecoveryPointEnum string

const (
	StartedRecoveryPoint       RecoveryPointEnum = "started"
	RideCreatedRecoveryPoint                     = "ride_created"
	ChargeCreatedRecoveryPoint                   = "charge_created"
	FinishedRecoveryPoint                        = "finished"
)

func (rp RecoveryPointEnum) String() string {
	return string(rp)
}
func (rp RecoveryPointEnum) IsValid() bool {
	switch rp {
	case StartedRecoveryPoint, RideCreatedRecoveryPoint,
		ChargeCreatedRecoveryPoint,
		FinishedRecoveryPoint:
		return true
	default:
		return false
	}
}

type RequestMethod string

func (m RequestMethod) IsValid() bool {
	switch m {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

func (m RequestMethod) String() string {
	return string(m)
}
