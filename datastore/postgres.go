package datastore

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite/stats"
	"github.com/SpirentOrion/trace"
	"github.com/lib/pq"
)

const (
	statPostgresExecSuffix         = ".exec"
	statPostgresExecLatencySuffix  = ".exec_latency"
	statPostgresQuerySuffix        = ".query"
	statPostgresQueryLatencySuffix = ".query_latency"
	statPostgresErrorSuffix        = ".error."
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

type PostgresDb struct {
	params      *PostgresParams
	logger      *log.Entry
	stats       stats.Stats
	statsPrefix string
	*sql.DB
}

func NewPostgresDb(params *PostgresParams, logger *log.Entry, stats stats.Stats) (*PostgresDb, error) {
	db, err := sql.Open(POSTGRES_PROVIDER, fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		params.User, params.Password, params.DbName, params.Host, params.Port, params.SslMode))
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(params.MaxIdleConns)
	db.SetMaxOpenConns(params.MaxOpenConns)

	return &PostgresDb{
		params:      params,
		logger:      logger,
		stats:       stats,
		statsPrefix: fmt.Sprintf("datastore.%s.%s.", POSTGRES_PROVIDER, params.DbName),
		DB:          db,
	}, nil
}

func (db *PostgresDb) String() string {
	return fmt.Sprintf("%s{%s:%d/%s}", POSTGRES_PROVIDER, db.params.Host, db.params.Port, db.params.DbName)
}

func (db *PostgresDb) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	var latency time.Duration

	s, _ := trace.Continue(POSTGRES_PROVIDER, db.String())
	trace.Run(s, func() {
		start := time.Now()
		res, err = db.DB.Exec(query, args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = "Exec"
			if err != nil {
				data["error"] = err
			} else {
				data["query"] = query
				rows, _ := res.RowsAffected()
				data["rows"] = rows
			}
		}
	})

	db.stats.Incr(db.statsPrefix+statPostgresExecSuffix, 1)
	db.stats.PrecisionTiming(db.statsPrefix+statPostgresExecLatencySuffix, latency)
	if err != nil {
		db.handleError("Exec", query, err)
	}
	return
}

func (db *PostgresDb) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	var latency time.Duration

	s, _ := trace.Continue(POSTGRES_PROVIDER, db.String())
	trace.Run(s, func() {
		start := time.Now()
		rows, err = db.DB.Query(query, args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = "Query"
			if err != nil {
				data["error"] = err
			} else {
				data["query"] = query
			}
		}
	})

	db.stats.Incr(db.statsPrefix+statPostgresQuerySuffix, 1)
	db.stats.PrecisionTiming(db.statsPrefix+statPostgresQueryLatencySuffix, latency)
	if err != nil {
		db.handleError("Query", query, err)
	}
	return
}

func (db *PostgresDb) handleError(op, query string, err error) {
	db.logger.WithFields(log.Fields{
		"provider": POSTGRES_PROVIDER,
		"user":     db.params.User,
		"dbname":   db.params.DbName,
		"host":     db.params.Host,
		"port":     db.params.Port,
		"op":       op,
		"query":    query,
		"error":    err,
	}).Error()

	pgErr, ok := err.(pq.Error)
	if ok {
		db.stats.Incr(db.statsPrefix+statPostgresErrorSuffix+string(pgErr.Code), 1)
	} else {
		db.stats.Incr(db.statsPrefix+statPostgresErrorSuffix+"other", 1)
	}
}
