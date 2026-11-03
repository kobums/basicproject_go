package services

// standin 은 테스트용 database/sql 드라이버다 (MariaDB 없이 생성 코드가 DB 까지 가는 경로를 확인한다).
//   - 모든 Exec / Query 를 기록한다 (어떤 SQL 과 값으로 DB 에 갔는지 검사).
//   - Query 결과는 테스트가 standinRows 로 정한다. 정하지 않았으면 에러 = "DB 없음" (생성 코드는 code "error").
//   - Exec 는 성공한다. insert 의 LastInsertId 는 standinInsertID 다.
// 컬럼은 위치로만 읽으므로(rows.Scan) 이름은 c0, c1 ... 이다.

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

const standinDriver = "standin"

const standinInsertID = 100

type standinCall struct {
	Exec  bool
	Query string
	Args  []driver.Value
}

var standin = struct {
	sync.Mutex
	calls []standinCall
	rows  func(query string, args []driver.Value) ([][]driver.Value, error)
}{}

var errStandinNoDB = errors.New("standin: no database")

func init() {
	sql.Register(standinDriver, standinDrv{})
}

// standinReset 는 기록을 지우고 Query 결과 함수를 바꾼다 (nil 이면 모든 Query 가 에러).
func standinReset(rows func(query string, args []driver.Value) ([][]driver.Value, error)) {
	standin.Lock()
	defer standin.Unlock()
	standin.calls = nil
	standin.rows = rows
}

func standinCalls() []standinCall {
	standin.Lock()
	defer standin.Unlock()
	return append([]standinCall(nil), standin.calls...)
}

// standinFind 는 query 에 sub 가 들어간 첫 기록을 돌려준다.
func standinFind(sub string) *standinCall {
	for _, c := range standinCalls() {
		if strings.Contains(c.Query, sub) {
			c := c
			return &c
		}
	}
	return nil
}

type standinDrv struct{}

func (standinDrv) Open(string) (driver.Conn, error) { return standinConn{}, nil }

type standinConn struct{}

func (standinConn) Prepare(query string) (driver.Stmt, error) { return standinStmt{query}, nil }
func (standinConn) Close() error                              { return nil }
func (standinConn) Begin() (driver.Tx, error)                 { return standinTx{}, nil }

type standinTx struct{}

func (standinTx) Commit() error   { return nil }
func (standinTx) Rollback() error { return nil }

type standinStmt struct{ query string }

func (s standinStmt) Close() error  { return nil }
func (s standinStmt) NumInput() int { return -1 }

func (s standinStmt) Exec(args []driver.Value) (driver.Result, error) {
	standin.Lock()
	defer standin.Unlock()
	standin.calls = append(standin.calls, standinCall{Exec: true, Query: s.query, Args: args})
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.query)), "insert") {
		return standinResult{}, nil
	}
	return driver.RowsAffected(1), nil
}

func (s standinStmt) Query(args []driver.Value) (driver.Rows, error) {
	standin.Lock()
	standin.calls = append(standin.calls, standinCall{Query: s.query, Args: args})
	fn := standin.rows
	standin.Unlock()
	if fn == nil {
		return nil, errStandinNoDB
	}
	rows, err := fn(s.query, args)
	if err != nil {
		return nil, err
	}
	return &standinRows{rows: rows}, nil
}

// standinResult 는 insert 의 LastInsertId 를 준다 (driver.RowsAffected 는 LastInsertId 가 에러다).
type standinResult struct{}

func (standinResult) LastInsertId() (int64, error) { return standinInsertID, nil }
func (standinResult) RowsAffected() (int64, error) { return 1, nil }

type standinRows struct {
	rows [][]driver.Value
	pos  int
}

func (r *standinRows) Columns() []string {
	n := 0
	if len(r.rows) > 0 {
		n = len(r.rows[0])
	}
	cols := make([]string, n)
	for i := range cols {
		cols[i] = fmt.Sprintf("c%d", i)
	}
	return cols
}

func (r *standinRows) Close() error { return nil }

func (r *standinRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}
