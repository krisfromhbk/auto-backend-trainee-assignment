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

// Config defines fields (with defaults) used for configuring http server and parsing them from environment variables
type Config struct {
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	Port uint16 `env:"PORT" envDefault:"9000"`
}

// WithConfig enables processing exported Config struct to acts as a source of config parameters for Server
func WithConfig(cfg Config) Option {
	return optionFunc(func(c *config) {
		c.addr = cfg.Host + ":" + strconv.FormatUint(uint64(cfg.Port), 10)
	})
}
