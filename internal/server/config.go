package server

import "strconv"

type Option interface {
	apply(*config)
}

type optionFunc func(c *config)

func (f optionFunc) apply(c *config) { f(c) }

// config defines fields used for configuring Server instance
type config struct {
	addr string
}

// EnvConfig defines fields used for parsing from environment variables
type EnvConfig struct {
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	Port uint16 `env:"PORT" envDefault:"9000"`
}

// WithEnvConfig enables processing exported EnvConfig struct to acts as a source of config parameters for Server
func WithEnvConfig(cfg EnvConfig) Option {
	return optionFunc(func(c *config) {
		c.addr = cfg.Host + ":" + strconv.FormatUint(uint64(cfg.Port), 10)
	})
}
