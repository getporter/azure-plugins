package config

type Config struct {
	// Env is the environment variable from which the connection string should be loaded.
	Env string
}

func New() Config {
	return Config{}
}
