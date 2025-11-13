package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	outboxmigrations "gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox/migrations"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type multiCloser struct {
	closers []io.Closer
}

func (m *multiCloser) Add(c io.Closer) {
	if c != nil {
		m.closers = append(m.closers, c)
	}
}

func (m *multiCloser) Close() error {
	var errs []error
	for _, c := range m.closers {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func newConnectionsContainer(
	config *config,
	logger logging.Logger,
	multiCloser *multiCloser,
) (container *connectionsContainer, err error) {
	containerBuilder := func() error {
		container = &connectionsContainer{}

		db, err := initMySQL(config)
		if err != nil {
			return fmt.Errorf("failed to init DB for migrations: %w", err)
		}
		multiCloser.Add(db)

		client := mysql.NewTransactionalClientFromSQLx(db)
		container.connectionPool = mysql.NewConnectionPool(client)
		container.client = client

		err = func() error {
			migrator, release, err := outboxmigrations.NewOutboxMigrator(context.Background(), container.connectionPool, logger, "domain")
			if err != nil {
				return err
			}
			defer release()
			err = migrator.Migrate()
			if err != nil {
				return err
			}
			if err = applyMigrations(db.DB, pathToMigrations); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
			logger.Info("Migrations applied successfully")
			return nil
		}()
		if err != nil {
			return err
		}

		// TODO: это конекшены к другим сервисам (в данном случае - gRPC)
		testConnection, err := grpc.NewClient(
			config.TestGRPCAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return err
		}

		multiCloser.Add(testConnection)
		container.testConnection = testConnection

		container.amqpConnection = amqp.NewAMQPConnection(appID, &amqp.ConnectionConfig{
			User:     config.AMQPUsername,
			Password: config.AMQPPassword,
			Host:     config.AMQPHost,
		}, logger)

		return nil
	}

	return container, containerBuilder()
}

type connectionsContainer struct {
	connectionPool mysql.ConnectionPool
	client         mysql.ClientContext
	amqpConnection amqp.Connection
	testConnection grpc.ClientConnInterface
}

func initMySQL(cfg *config) (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("mysql", cfg.buildDSN())
	if err != nil || db == nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxConn)

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	return db, nil
}
