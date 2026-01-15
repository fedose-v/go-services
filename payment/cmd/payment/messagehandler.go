package main

import (
	"errors"
	"net/http"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"payment/pkg/payment/infrastructure/consumer"
	"payment/pkg/payment/infrastructure/integrationevent"
	inframysql "payment/pkg/payment/infrastructure/mysql"
)

type messageHandlerConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	AMQP     AMQP     `envconfig:"amqp" required:"true"`
}

func messageHandler(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "message-handler",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[messageHandlerConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				err = errors.Join(err, closer.Close())
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)
			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			amqpConnection := newAMQPConnection(cnf.AMQP, logger)
			amqpEventProducer := amqpConnection.Producer(
				&amqp.ExchangeConfig{
					Name:    integrationevent.ExchangeName,
					Kind:    integrationevent.ExchangeKind,
					Durable: true,
				},
				nil,
				nil,
			)

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			eventDispatcher := outbox.NewEventDispatcher(appID, integrationevent.TransportName, integrationevent.NewEventSerializer(), libUoW)

			eventConsumer, err := consumer.NewEventConsumer(c.Context, amqpConnection, databaseConnectionPool, logger, eventDispatcher)
			if err != nil {
				return err
			}

			queueConfig := &amqp.QueueConfig{
				Name:    "payment_events",
				Durable: true,
			}
			bindConfig := &amqp.BindConfig{
				QueueName:    "payment_events",
				ExchangeName: "domain_event_exchange",
				RoutingKeys:  []string{"user.*"},
			}
			amqpConnection.Consumer(
				c.Context,
				eventConsumer.Handler(),
				queueConfig,
				bindConfig,
				nil,
			)
			err = amqpConnection.Start()
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				return amqpConnection.Stop()
			}))

			outboxEventHandler := outbox.NewEventHandler(outbox.EventHandlerConfig{
				TransportName:  integrationevent.TransportName,
				Transport:      integrationevent.NewTransport(logger, amqpEventProducer),
				ConnectionPool: databaseConnectionPool,
				Logger:         logger,
			})

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				return outboxEventHandler.Start(c.Context)
			})

			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				// nolint:gosec
				server := http.Server{
					Addr:    cnf.Service.HTTPAddress,
					Handler: router,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
