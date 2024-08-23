package rides

type RideService interface {
	CreateRide(ride Ride)
	UpdateRide(ride Ride)
}

type rideService struct {
}
