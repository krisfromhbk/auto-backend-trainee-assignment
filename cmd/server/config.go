package main

import (
	"auto/internal/server"
	"errors"
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"os"
)

var helpErr = errors.New("help has been required")

type config struct {
	http *server.Config
}

type options struct {
	logger *zap.Logger
	*config
}

func (o options) installServerFlags(flags *pflag.FlagSet) {
	o.logger.Debug("installing config flags")
	flags.StringVar(&o.config.http.Host, "host", o.config.http.Host, "Application host")
	flags.Uint16Var(&o.config.http.Port, "port", o.config.http.Port, "Application port")
}

func newConfig(logger *zap.Logger) (config, error) {
	opts := options{
		logger: logger,
		config: &config{
			http: &server.Config{},
		},
	}

	if err := env.Parse(opts.config.http); err != nil {
		logger.Error("parsing server environment config", zap.Error(err))
	}

	serverFlags := pflag.NewFlagSet("http_server", pflag.ContinueOnError)
	opts.installServerFlags(serverFlags)

	if err := serverFlags.Parse(os.Args[1:]); err != nil {
		if errors.As(err, &pflag.ErrHelp) {
			//logger.Info("", zap.String("usage", serverFlags.FlagUsages()))
			return config{}, helpErr
		}
		logger.Error("can not parse http server flags", zap.Error(err))
		return config{}, err
	}

	return *opts.config, nil
}
