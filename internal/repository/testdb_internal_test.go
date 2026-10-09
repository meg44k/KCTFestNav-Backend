package repository

import (
	"context"
	"database/sql"
)

// 外部キーのチェックを切って表を空にする。SET は接続ごとの設定なので、
// プールの別の接続で TRUNCATE されないよう 1 つの接続でまとめて流す
func truncateTables(db *sql.DB, tables ...string) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}
	defer conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1")
	for _, t := range tables {
		if _, err := conn.ExecContext(ctx, "TRUNCATE TABLE "+t); err != nil {
			return err
		}
	}
	return nil
}
