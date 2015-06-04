package datastore

import (
	"database/sql"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite/stats"
	"github.com/SpirentOrion/trace"
)

const (
	statSqlExecSuffix         = ".exec"
	statSqlExecLatencySuffix  = ".exec_latency"
	statSqlQuerySuffix        = ".query"
	statSqlQueryLatencySuffix = ".query_latency"
	statSqlTxSuffix           = ".tx"
	statSqlErrorSuffix        = ".error."
)

type SqlDb struct {
	provider         string
	name             string
	logger           *log.Entry
	stats            stats.Stats
	statsPrefix      string
	handleError      func(db *SqlDb, op, query string, err error)
	shouldRetryError func(db *SqlDb, err error) bool
	*sql.DB
}

func (db *SqlDb) String() string {
	return db.name
}

func (db *SqlDb) Begin() (*SqlTx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	db.stats.Incr(db.statsPrefix+statSqlTxSuffix, 1)
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
			if err != nil {
				data["error"] = err
			} else {
				data["query"] = query
				rows, _ := res.RowsAffected()
				data["rows"] = rows
			}
		}
	})

	db.stats.Incr(db.statsPrefix+statSqlExecSuffix, 1)
	db.stats.PrecisionTiming(db.statsPrefix+statSqlExecLatencySuffix, latency)
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
			if err != nil {
				data["error"] = err
			} else {
				data["query"] = query
			}
		}
	})

	db.stats.Incr(db.statsPrefix+statSqlQuerySuffix, 1)
	db.stats.PrecisionTiming(db.statsPrefix+statSqlQueryLatencySuffix, latency)
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
			if err != nil {
				data["error"] = err
			} else {
				data["query"] = query
				rows, _ := res.RowsAffected()
				data["rows"] = rows
			}
		}
	})

	tx.db.stats.Incr(tx.db.statsPrefix+statSqlExecSuffix, 1)
	tx.db.stats.PrecisionTiming(tx.db.statsPrefix+statSqlExecLatencySuffix, latency)
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
			if err != nil {
				data["error"] = err
			} else {
				data["query"] = query
			}
		}
	})

	tx.db.stats.Incr(tx.db.statsPrefix+statSqlQuerySuffix, 1)
	tx.db.stats.PrecisionTiming(tx.db.statsPrefix+statSqlQueryLatencySuffix, latency)
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

	stmt.db.stats.Incr(stmt.db.statsPrefix+statSqlExecSuffix, 1)
	stmt.db.stats.PrecisionTiming(stmt.db.statsPrefix+statSqlExecLatencySuffix, latency)
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

	stmt.db.stats.Incr(stmt.db.statsPrefix+statSqlQuerySuffix, 1)
	stmt.db.stats.PrecisionTiming(stmt.db.statsPrefix+statSqlQueryLatencySuffix, latency)
	if err != nil {
		stmt.db.handleError(stmt.db, op, "", err)
	}
	return
}

func (s *SqlStmt) Close() error {
	return s.Stmt.Close()
}
