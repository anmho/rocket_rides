package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_makeConnString(t *testing.T) {
	user := "admin"
	pass := "admin"
	port := "5433"
	host := "localhost"
	name := "rocket_rides"

	connStr := makeConnString(user, pass, host, name, port)

	assert.Equal(t, "postgres://admin:admin@localhost:5433/rocket_rides", connStr)
}
