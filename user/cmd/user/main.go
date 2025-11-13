package main

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	applogger "gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/logging"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

const appID = "user"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cnf, err := parseEnv()
	if err != nil {
		stdlog.Fatal(err)
	}

	logger := logging.NewJSONLogger(&logging.Config{AppName: appID})

	err = runApp(ctx, cnf, logger)
	switch errors.Cause(err) {
	case nil:
		logger.Info("service finished")
	default:
		logger.FatalError(err)
	}
}

func runApp(
	ctx context.Context,
	config *config,
	logger applogger.Logger,
) (err error) {
	closer := &multiCloser{}
	defer func() {
		if closeErr := closer.Close(); closeErr != nil {
			err = errors.Wrap(err, closeErr.Error())
			if err == nil {
				err = closeErr
			}
		}
	}()

	app := cli.App{
		Name: appID,
		Commands: []*cli.Command{
			service(config, logger, closer),
		},
	}

	return app.RunContext(ctx, os.Args)
}
