package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetString(key, fallback string) string {

	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}
	return val

}

func GetInt(key string, fallback int) int {

	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	valInt, err := strconv.Atoi(val)

	if err != nil {
		fmt.Printf("Error, %v", err)
		return fallback
	}

	return valInt

}

func GetDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * 24 * time.Hour
	}
	// On parse error, use fallback
	return fallback
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	valBool, err := strconv.ParseBool(val)

	if err != nil {
		fmt.Printf("Error, %v", err)
		return fallback
	}

	return valBool
}
