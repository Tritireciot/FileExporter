package pgutils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PGXSqlConn struct {
	m_ctx    context.Context
	m_conn   *pgxpool.Conn
	m_cancel context.CancelFunc
}

type PGXSQLLocker struct {
	sql_conn PGXSqlConn
}

func CreatePGXSQLLocker(pool *pgxpool.Pool) (*PGXSQLLocker, error) {
	var err error
	var locker PGXSQLLocker
	if pool == nil {
		return nil, errors.New("empty driver pointer initialization")
	}
	if locker.sql_conn.m_ctx == nil {
		locker.sql_conn.m_ctx, locker.sql_conn.m_cancel = context.WithCancel(context.Background())
	}
	if locker.sql_conn.m_conn == nil {
		acquireCtx, acquireCancel := context.WithTimeout(locker.sql_conn.m_ctx, 3*time.Second)
		defer acquireCancel()
		if locker.sql_conn.m_conn, err = pool.Acquire(acquireCtx); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return &locker, nil
}

func (locker *PGXSQLLocker) Close() {
	locker.sql_conn.m_conn.Conn().Close(locker.sql_conn.m_ctx)
	locker.sql_conn.m_cancel()
}

/*
pg_advisory_unlock(bigint)
*/
func (locker *PGXSQLLocker) ReleaseLock(lock_key string) error {
	var err error
	if locker == nil {
		return fmt.Errorf("%v", "locker not created")
	}
	if locker.sql_conn.m_conn == nil {
		return fmt.Errorf("%v", "db connection object is null")
	}

	_, err = locker.sql_conn.m_conn.Exec(locker.sql_conn.m_ctx, `SELECT pg_advisory_unlock(('x' || md5($1))::bit(32)::int)`, lock_key)

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// Попытка заблокировать ресурс с ключом lock_key.
//
//	В случае если мы уже заблокировали ресурс - возврат УСПЕХ
func (locker *PGXSQLLocker) TryLock(lock_key string) error {
	var err error
	if locker == nil {
		return fmt.Errorf("%v", "locker not created")
	}

	if locker.sql_conn.m_conn == nil {
		return fmt.Errorf("%v", "db connection object is null")
	}

	execCtx, execCancel := context.WithTimeout(locker.sql_conn.m_ctx, 3*time.Second)
	defer execCancel()

	var unlocked bool
	err = locker.sql_conn.m_conn.QueryRow(execCtx, `SELECT pg_try_advisory_lock(('x' || md5($1))::bit(32)::int);`, lock_key).Scan(&unlocked)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if !unlocked {
		return fmt.Errorf("already locked by another instance")
	}

	return nil
}

// Проверка на состояние ресурса с ключом lock_key
func (locker *PGXSQLLocker) TestLock(lock_key string) (result int, err error) {

	if locker == nil {
		return LockResult_error, fmt.Errorf("%v", "locker not created")
	}
	if locker.sql_conn.m_conn == nil {
		return LockResult_error, fmt.Errorf("%v", "db connection object is null")
	}

	pingCtx, pingCancel := context.WithTimeout(locker.sql_conn.m_ctx, 3*time.Second)
	defer pingCancel()

	if err = locker.sql_conn.m_conn.Ping(pingCtx); err != nil {
		return LockResult_error, fmt.Errorf("db connection lost: %v", err)
	}

	lockAcquired := false
	err = locker.sql_conn.m_conn.QueryRow(pingCtx,
		`SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype = 'advisory' AND objid = ('x' || md5($1))::bit(32)::int);`, lock_key).Scan(&lockAcquired)
	if err != nil {
		return LockResult_error, fmt.Errorf("failed to check lock state: %v", err)
	}

	if !lockAcquired {
		return LockResult_another, nil
	}

	return LockResult_self, nil
}
