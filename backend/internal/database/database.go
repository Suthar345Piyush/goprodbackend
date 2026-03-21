package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/Suthar345Piyush/goprodbackend/internal/config"
	loggerConfig "github.com/Suthar345Piyush/goprodbackend/internal/logger"
	pgxzero "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/newrelic/go-agent/v3/integrations/nrpgx5"
	"github.com/rs/zerolog"
)

type Database struct {
	Pool *pgxpool.Pool
	log  *zerolog.Logger
}

// we have multiple tracers like for dev and production both , so the struct contains an slice of mulitple tracers

type multiTracer struct {
	tracers []any
}

// trace query start and end , pgx tracer interface  and it will return the context

func (mt *multiTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {

	for _, tracer := range mt.tracers {
		if t, ok := tracer.(interface {
			TraceQueryStart(context.Context, *pgx.Conn, pgx.TraceQueryStartData) context.Context
		}); ok {
			ctx = t.TraceQueryStart(ctx, conn, data)
		}
	}

	return ctx
}

// same for trace query end

func (mt *multiTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {

	for _, tracer := range mt.tracers {
		if t, ok := tracer.(interface {
			TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData)
		}); ok {
			t.TraceQueryEnd(ctx, conn, data)
		}
	}
}

const DatabasePingTimeout = 10

// main function

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerConfig.LoggerService) (*Database, error) {

	// host:port using JoinHostPort
	// port - integer  -> string

	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// encode the password

	encodedPassword := url.QueryEscape(cfg.Database.Password)

	// database(user) , hostport , password , db name , db sslmode

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// making pool of db connections

	pgxPoolConfig, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
	}

	// adding the newRelic PostgresQl instrumentation

	if loggerService != nil && loggerService.GetApplication() != nil {
		pgxPoolConfig.ConnConfig.Tracer = nrpgx5.NewTracer()
	}

	// checking the config env

	if cfg.Primary.Env == "local" {

		globalLevel := logger.GetLevel()

		pgxLogger := loggerConfig.NewPgxLogger(globalLevel)

		// chain of tracers - on first new relic and then local logging

		if pgxPoolConfig.ConnConfig.Tracer != nil {

			// if new relic tracers exists , then creating multi tracers

			localTracer := &tracelog.TraceLog{
				Logger: pgxzero.NewLogger(pgxLogger),

				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLevel)),
			}

			// filling that tracers slice

			pgxPoolConfig.ConnConfig.Tracer = &multiTracer{
				tracers: []any{pgxPoolConfig.ConnConfig.Tracer, localTracer},
			}
		} else {

			pgxPoolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLevel)),
			}

		}
	}

	// creating the new connection pool

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	database := &Database{
		Pool: pool,
		log:  logger,
	}

	// logging part of the database connection verification

	ctx, cancel := context.WithTimeout(context.Background(), DatabasePingTimeout*time.Second)

	defer cancel()

	// ping usecase - ping takes the connection from pool , then run a sql statement against it , if sql return without error means db ping is successfull , otherwise error retruns

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// if everything works well

	logger.Info().Msg("connected to database")

	return database, nil

}

// at last we have to close the db connection pool

func (db *Database) Close() error {
	db.log.Info().Msg("closing database connection pool")
	db.Pool.Close()

	return nil
}
