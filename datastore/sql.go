package datastore

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/trace"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	sqlOps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sql_operations_total",
			Help: "How many SQL operations occurred, partitioned by host and database.",
		},
		[]string{"host", "database"},
	)

	sqlOpLatencies = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "sql_operation_latency_milliseconds",
			Help: "SQL operation latencies in milliseconds, partitioned by host and database.",
		},
		[]string{"host", "database"},
	)

	sqlErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sql_errors_total",
			Help: "How many SQL errors occurred, partitioned by host, database, and error code.",
		},
		[]string{"host", "database", "error_code"},
	)
)

func init() {
	prometheus.MustRegister(sqlOps)
	prometheus.MustRegister(sqlOpLatencies)
	prometheus.MustRegister(sqlErrors)
}

type SqlDb struct {
	provider         string
	host             string
	name             string
	logger           *log.Logger
	handleError      func(db *SqlDb, op, query string, err error)
	shouldRetryError func(db *SqlDb, err error) bool
	*sql.DB
}

func (db *SqlDb) String() string {
	return fmt.Sprintf("%s/%s", db.host, db.name)
}

func (db *SqlDb) Begin() (*SqlTx, error) {
	start := time.Now()
	tx, err := db.DB.Begin()
	latency := time.Since(start)

	sqlOps.WithLabelValues(db.host, db.name).Inc()
	sqlOpLatencies.WithLabelValues(db.host, db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		return nil, err
	}

	return &SqlTx{
		db: db,
		Tx: tx,
	}, nil
}

func (db *SqlDb) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	var latency time.Duration
	const op = "Exec"

	s, _ := trace.Continue(db.provider, db.String())
	trace.Run(s, func() {
		start := time.Now()
		res, err = db.DB.Exec(query, args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			data["query"] = query
			if err != nil {
				data["error"] = err
			} else {
				rows, _ := res.RowsAffected()
				data["rows"] = rows
			}
		}
	})

	sqlOps.WithLabelValues(db.host, db.name).Inc()
	sqlOpLatencies.WithLabelValues(db.host, db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		db.handleError(db, op, query, err)
	}
	return
}

func (db *SqlDb) Prepare(query string) (*SqlStmt, error) {
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &SqlStmt{
		db:   db,
		Stmt: stmt,
	}, nil
}

func (db *SqlDb) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	var latency time.Duration
	const op = "Query"

	s, _ := trace.Continue(db.provider, db.String())
	trace.Run(s, func() {
		start := time.Now()
		rows, err = db.DB.Query(query, args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			data["query"] = query
			if err != nil {
				data["error"] = err
			}
		}
	})

	sqlOps.WithLabelValues(db.host, db.name).Inc()
	sqlOpLatencies.WithLabelValues(db.host, db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		db.handleError(db, op, query, err)
	}
	return
}

func (db *SqlDb) ShouldRetryError(err error) bool {
	return db.shouldRetryError(db, err)
}

type SqlTx struct {
	db *SqlDb
	*sql.Tx
}

func (tx *SqlTx) Commit() (err error) {
	var latency time.Duration
	const op = "Commit"

	s, _ := trace.Continue(tx.db.provider, tx.db.String())
	trace.Run(s, func() {
		start := time.Now()
		err = tx.Tx.Commit()
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			if err != nil {
				data["error"] = err
			}
		}
	})

	sqlOps.WithLabelValues(tx.db.host, tx.db.name).Inc()
	sqlOpLatencies.WithLabelValues(tx.db.host, tx.db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		if tx.db.shouldRetryError(tx.db, err) {
			tx.Tx.Rollback()
		}
		tx.db.handleError(tx.db, op, "", err)
	}
	return
}

func (tx *SqlTx) Rollback() error {
	return tx.Tx.Rollback()
}

func (tx *SqlTx) Stmt(stmt *SqlStmt) *SqlStmt {
	return &SqlStmt{
		db:   tx.db,
		Stmt: tx.Tx.Stmt(stmt.Stmt),
	}
}

func (tx *SqlTx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	var latency time.Duration
	const op = "Exec"

	s, _ := trace.Continue(tx.db.provider, tx.db.String())
	trace.Run(s, func() {
		start := time.Now()
		res, err = tx.Tx.Exec(query, args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			data["query"] = query
			if err != nil {
				data["error"] = err
			} else {
				rows, _ := res.RowsAffected()
				data["rows"] = rows
			}
		}
	})

	sqlOps.WithLabelValues(tx.db.host, tx.db.name).Inc()
	sqlOpLatencies.WithLabelValues(tx.db.host, tx.db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		if tx.db.shouldRetryError(tx.db, err) {
			tx.Tx.Rollback()
		}
		tx.db.handleError(tx.db, op, query, err)
	}
	return
}

func (tx *SqlTx) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	var latency time.Duration
	const op = "Query"

	s, _ := trace.Continue(tx.db.provider, tx.db.String())
	trace.Run(s, func() {
		start := time.Now()
		rows, err = tx.Tx.Query(query, args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			data["query"] = query
			if err != nil {
				data["error"] = err
			}
		}
	})

	sqlOps.WithLabelValues(tx.db.host, tx.db.name).Inc()
	sqlOpLatencies.WithLabelValues(tx.db.host, tx.db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		if tx.db.shouldRetryError(tx.db, err) {
			tx.Tx.Rollback()
		}
		tx.db.handleError(tx.db, op, query, err)
	}
	return
}

type SqlStmt struct {
	db *SqlDb
	*sql.Stmt
}

func (stmt *SqlStmt) Exec(args ...interface{}) (res sql.Result, err error) {
	var latency time.Duration
	const op = "StmtExec"

	s, _ := trace.Continue(stmt.db.provider, stmt.db.String())
	trace.Run(s, func() {
		start := time.Now()
		res, err = stmt.Stmt.Exec(args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			if err != nil {
				data["error"] = err
			} else {
				rows, _ := res.RowsAffected()
				data["rows"] = rows
			}
		}
	})

	sqlOps.WithLabelValues(stmt.db.host, stmt.db.name).Inc()
	sqlOpLatencies.WithLabelValues(stmt.db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		stmt.db.handleError(stmt.db, op, "", err)
	}
	return
}

func (stmt *SqlStmt) Query(args ...interface{}) (rows *sql.Rows, err error) {
	var latency time.Duration
	const op = "StmtQuery"

	s, _ := trace.Continue(stmt.db.provider, stmt.db.String())
	trace.Run(s, func() {
		start := time.Now()
		rows, err = stmt.Stmt.Query(args...)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			if err != nil {
				data["error"] = err
			}
		}
	})

	sqlOps.WithLabelValues(stmt.db.host, stmt.db.name).Inc()
	sqlOpLatencies.WithLabelValues(stmt.db.name).Observe(latency.Seconds() / 1000)
	if err != nil {
		stmt.db.handleError(stmt.db, op, "", err)
	}
	return
}

func (s *SqlStmt) Close() error {
	return s.Stmt.Close()
}
