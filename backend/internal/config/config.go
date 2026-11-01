// Package config loads the application configuration from the environment.
package config

import "errors"

// Config holds the settings the API needs at startup.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

// Load reads the configuration through getenv, which has the same signature
// as os.LookupEnv so tests can pass a fake environment.
func Load(getenv func(string) (string, bool)) (Config, error) {
	cfg := Config{HTTPAddr: ":8080"}
	var errs []error

	if v, ok := getenv("HTTP_ADDR"); ok && v != "" {
		cfg.HTTPAddr = v
	}

	v, ok := getenv("DATABASE_URL")
	if !ok || v == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	cfg.DatabaseURL = v

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
