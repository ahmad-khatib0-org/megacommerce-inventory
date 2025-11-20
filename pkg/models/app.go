// Package models contains models for inventory, config ...
package models

type Config struct {
	Service Service `mapstructure:"service"`
}

type Service struct {
	Env                  string `mapstructure:"env"`
	GrpcURL              string `mapstructure:"grpc_url"`
	CommonServiceGrpcURL string `mapstructure:"common_service_grpc_url"`
}

