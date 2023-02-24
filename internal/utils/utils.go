package utils

import (
	"os"

	"github.com/joho/godotenv"
)

func getEnvironment() string {
	_, exists := os.LookupEnv("APP_ENV")
	if !exists {
		godotenv.Load()
	}

	env := os.Getenv("APP_ENV")
	if len(env) == 0 {
		return "local"
	}

	return env
}

// Check if the application is running in production environment.
func IsProduction() bool {
	return getEnvironment() == "production"
}

// Check if the application is running in development environment.
func IsDevelopment() bool {
	return !IsProduction()
}

// Check if the application is running in local development environment.
func IsLocal() bool {
	return getEnvironment() == "local"
}

// Get the name of the current application environment.
func GetEnvName() string {
	return getEnvironment()
}
