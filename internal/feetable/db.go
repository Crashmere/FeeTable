package feetable

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	_ "modernc.org/sqlite"
	"net/url"
	"os"
	"path/filepath"
)

//go:embed schema.sql
var schema string

type Store struct{ db *sql.DB }

func Open(path string, create bool) (*Store, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if create {
		if err = os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
			return nil, err
		}
		f, e := os.OpenFile(absolute, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return nil, e
		}
		if e = f.Close(); e != nil {
			return nil, e
		}
	} else if info, e := os.Stat(absolute); e != nil {
		return nil, e
	} else if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, fmt.Errorf("数据库为空或不是普通文件")
	}
	u := url.URL{Scheme: "file", Path: absolute, RawQuery: url.Values{"mode": {"rw"}, "_pragma": {"foreign_keys(1)", "busy_timeout(5000)", "journal_mode(WAL)", "synchronous(FULL)"}}.Encode()}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &Store{db: db}
	ctx := context.Background()
	if create {
		err = s.transaction(ctx, func(tx *sql.Tx) error { _, e := tx.ExecContext(ctx, schema); return e })
	} else {
		err = checkDB(ctx, db)
	}
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error                   { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *Store) RequireCurrentSchema(ctx context.Context) error {
	var version int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version != 2 {
		return fmt.Errorf("数据库需要升级，请先停止服务、备份并运行 feetable migrate")
	}
	return nil
}
func (s *Store) transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
func checkDB(ctx context.Context, db queryer) error {
	var version, appID int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx, "PRAGMA application_id").Scan(&appID); err != nil {
		return err
	}
	if (version != 1 && version != 2) || appID != 1179931714 {
		return fmt.Errorf("不支持的 FeeTable 数据库")
	}
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("数据库完整性检查失败")
	}
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return fmt.Errorf("数据库存在失效外键")
	}
	return rows.Err()
}
