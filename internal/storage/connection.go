package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"order_notification_system/cmd/config"
)

func NewConnectionPool(ctx context.Context, conf *config.Config) (*pgxpool.Pool, error) {
	dbConf := conf.Databases["default"]
	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConf.User, dbConf.Pwd, dbConf.Host, dbConf.Port, dbConf.Name,
	)
	poolConfig, err := pgxpool.ParseConfig(pgURL)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres URI: %v", err.Error())
	}

	poolConfig.MaxConns = int32(conf.Pool.MaxConn)
	poolConfig.MinConns = int32(conf.Pool.MinConn)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	logrus.WithField("max connections", conf.Pool.MaxConn).Info("started connection pool")

	return pool, pool.Ping(ctx)
}
