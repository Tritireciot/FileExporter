package pgutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type TSqlConn struct {
	m_db   *sql.DB
	m_ctx  context.Context
	m_conn *sql.Conn
}

type ISqlLocker interface {
	GetLock(lock_key string) error
	ReleaseLock(lock_key string) error
	TestLock(lock_key string) error
}

type PSQLLocker struct {
	sql_conn TSqlConn
}

func CreatePSQLLocker(db *sql.DB) (*PSQLLocker, error) {
	var err error
	var locker PSQLLocker
	if db == nil {
		return nil, errors.New("empty driver pointer initialization")
	}
	if locker.sql_conn.m_db == nil {
		locker.sql_conn.m_db = db
	}
	if locker.sql_conn.m_ctx == nil {
		locker.sql_conn.m_ctx = context.Background()
	}
	if locker.sql_conn.m_conn == nil {
		locker.sql_conn.m_conn, err = db.Conn(locker.sql_conn.m_ctx)
	}
	if err != nil {
		return nil, err
	}
	return &locker, nil
}

func Close(locker *PSQLLocker) {
	var conn TSqlConn
	if locker == nil {
		return
	}
	conn = locker.sql_conn
	if conn.m_conn != nil {
		conn.m_conn.Close()
	}
	conn.m_ctx.Done()
}

/*
	pg_advisory_lock(bigint)
*/

func (locker *PSQLLocker) GetLock(lock_key string) error {
	var err error
	if locker == nil {
		return fmt.Errorf("%v", "locker not created")
	}
	if locker.sql_conn.m_db == nil {
		return fmt.Errorf("%v", "database object is null")
	}

	if locker.sql_conn.m_conn == nil {
		return fmt.Errorf("%v", "db connection object is null")
	}

	_, err = locker.sql_conn.m_conn.ExecContext(locker.sql_conn.m_ctx, `SELECT pg_try_advisory_lock(('x' || md5(?))::bit(32)::int);`, lock_key)
	if err != nil {
		fmt.Println(err)

		return err
	}

	return nil
}

/*
pg_advisory_unlock(bigint)
*/
func (locker *PSQLLocker) ReleaseLock(lock_key string) error {
	var err error
	if locker == nil {
		return fmt.Errorf("%v", "locker not created")
	}
	if locker.sql_conn.m_db == nil {
		return fmt.Errorf("%v", "database object is null")
	}

	if locker.sql_conn.m_conn == nil {
		return fmt.Errorf("%v", "db connection object is null")
	}

	_, err = locker.sql_conn.m_conn.ExecContext(locker.sql_conn.m_ctx, `SELECT pg_advisory_unlock(('x' || md5(?))::bit(32)::int)`, lock_key)

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

const (
	LockResult_error   = -1
	LockResult_empty   = 0
	LockResult_self    = 1
	LockResult_another = 2
)

// Попытка заблокировать ресурс с ключом lock_key.
//
//	В случае если мы уже заблокировали ресурс - возврат УСПЕХ
func (locker *PSQLLocker) TryLock(lock_key string) error {
	if res, err := locker.TestLock(lock_key); err != nil {
		return fmt.Errorf("can't lock: %v", err)
	} else {
		switch res {
		case LockResult_empty:
			if err = locker.GetLock(lock_key); err != nil {
				return err
			}
		case LockResult_another:
			return fmt.Errorf("already locked by another")
		case LockResult_error:
			return fmt.Errorf("unexpected lock error")
		}
	}
	return nil
}

// Проверка на состояние ресурса с ключом lock_key
func (locker *PSQLLocker) TestLock(lock_key string) (result int, err error) {

	if locker == nil {
		return LockResult_error, fmt.Errorf("%v", "locker not created")
	}
	if locker.sql_conn.m_db == nil {
		return LockResult_error, fmt.Errorf("%v", "database object is null")
	}

	if locker.sql_conn.m_conn == nil {
		return LockResult_error, fmt.Errorf("%v", "db connection object is null")
	}

	rows, err := locker.sql_conn.m_conn.QueryContext(locker.sql_conn.m_ctx,
		`SELECT * FROM pg_locks WHERE locktype = 'advisory' and objid = ('x' || md5(?))::bit(32)::int;`, lock_key)
	if err != nil {
		fmt.Println(err)
		return LockResult_error, err
	}
	defer rows.Close()
	if rows.Next() {
		var lock_int int
		rows.Scan(&lock_int)
		if lock_int == 0 {
			result = LockResult_self
		}
		if lock_int == 1 {
			result = LockResult_another
		}
	} else {
		result = LockResult_empty
	}
	return
}
