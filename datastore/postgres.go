package datastore

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	log "github.com/SpirentOrion/logrus"
	"github.com/lib/pq"
)

const (
	PostgresErrorSerializationFailure = "40001"
)

// PostgresParams holds connection and auth properties for
// Postgres-based datastores.
type PostgresParams struct {
	User         string
	Password     string
	DbName       string
	Host         string
	Port         int
	SslMode      string
	MaxIdleConns int
	MaxOpenConns int
}

// NewPostgresParams extracts Progres provider parameters from a
// generic string map and returns a PostgresParams structure.
func NewPostgresParams(params map[string]string) (*PostgresParams, error) {
	p := &PostgresParams{
		User:     params["user"],
		Password: params["password"],
		DbName:   params["db_name"],
		Host:     params["host"],
		Port:     5432,
		SslMode:  params["ssl_mode"],
	}

	if p.User == "" {
		return nil, errors.New("Postgres providers require a 'user' parameter")
	}
	if p.Password == "" {
		return nil, errors.New("Postgres providers require a 'password' parameter")
	}
	if p.DbName == "" {
		return nil, errors.New("Postgres providers require a 'db_name' parameter")
	}
	if p.Host == "" {
		return nil, errors.New("Postgres providers require a 'host' parameter")
	}
	if port, err := strconv.Atoi(params["port"]); err == nil {
		p.Port = port
	}
	if p.SslMode == "" {
		p.SslMode = "require"
	}
	if maxIdleConns, err := strconv.Atoi(params["max_idle_conns"]); err == nil {
		p.MaxIdleConns = maxIdleConns
	}
	if maxOpenConns, err := strconv.Atoi(params["max_open_conns"]); err == nil {
		p.MaxOpenConns = maxOpenConns
	}

	return p, nil
}

func NewPostgresDb(params *PostgresParams, logger *log.Logger) (*SqlDb, error) {
	db, err := sql.Open(POSTGRES_PROVIDER, fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		params.User, params.Password, params.DbName, params.Host, params.Port, params.SslMode))
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(params.MaxIdleConns)
	db.SetMaxOpenConns(params.MaxOpenConns)

	return &SqlDb{
		provider:         POSTGRES_PROVIDER,
		host:             params.Host,
		name:             params.DbName,
		logger:           logger,
		handleError:      handlePostgresError,
		shouldRetryError: shouldRetryPostgresError,
		DB:               db,
	}, nil
}

func handlePostgresError(db *SqlDb, op, query string, err error) {
	pgErr, ok := err.(*pq.Error)
	if ok {
		db.logger.WithFields(log.Fields{
			"provider":       db.provider,
			"name":           db.name,
			"op":             op,
			"query":          query,
			"severity":       pgErr.Severity,
			"code":           pgErr.Code,
			"detail":         pgErr.Detail,
			"hint":           pgErr.Hint,
			"position":       pgErr.Position,
			"where":          pgErr.Where,
			"schema":         pgErr.Schema,
			"table":          pgErr.Table,
			"column":         pgErr.Column,
			"data_type_name": pgErr.DataTypeName,
			"constraint":     pgErr.Constraint,
			"file":           pgErr.File,
			"line":           pgErr.Line,
			"routine":        pgErr.Routine,
		}).Error(pgErr.Message)

		sqlErrors.WithLabelValues(db.host, db.name, string(pgErr.Code)).Inc()
	} else {
		db.logger.WithFields(log.Fields{
			"provider": db.provider,
			"name":     db.name,
			"op":       op,
			"query":    query,
		}).Error(err.Error())
	}
}

func shouldRetryPostgresError(db *SqlDb, err error) bool {
	if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == PostgresErrorSerializationFailure {
		return true
	}
	return false
}
