package env

import (
	"fmt"
	"os"
)

const (
	DatabaseURL  = "DATABASE_URL"
	RestHost     = "REST_HOST"
	ProtocolHost = "PROTOCOL_HOST"
)

func NewErrNotSet(env string) error {
	return fmt.Errorf("env %s isn't set", env)
}

func Get(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", NewErrNotSet(key)
	}
	return value, nil
}

func GetOptional(key string, optional string) string {
	value := os.Getenv(key)
	if value == "" {
		return optional
	}
	return value
}
