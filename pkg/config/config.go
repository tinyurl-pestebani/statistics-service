package config

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"strconv"
)

type ServiceConfig struct {
	Port int
}

func NewServiceConfig(port int) (*ServiceConfig, error) {
	return &ServiceConfig{
		Port: port,
	}, nil
}

func NewServiceConfigFromEnv() (*ServiceConfig, error) {
	port := 8082
	var err error
	portSt := os.Getenv("STATISTICS_SERVICE_PORT")

	if portSt != "" {
		port, err = strconv.Atoi(portSt)
	}

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return NewServiceConfig(port)

}
