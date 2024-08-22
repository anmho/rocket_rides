package idempotency

import "net/http"

type RecoveryPointEnum string

const (
	RecoveryPointStarted       RecoveryPointEnum = "started"
	RecoveryPointRideCreated                     = "ride_created"
	RecoveryPointChargeStarted                   = "charge_started"
	RecoveryPointFinished                        = "finished"
)

func (rp RecoveryPointEnum) IsValid() bool {
	switch rp {
	case RecoveryPointStarted, RecoveryPointRideCreated,
		RecoveryPointChargeStarted,
		RecoveryPointFinished:
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
