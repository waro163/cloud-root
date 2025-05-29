package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type BaseRepo struct {
	DB *sqlx.DB
}

func (repo *BaseRepo) Transaction(ctx context.Context, opts *sql.TxOptions, funs ...func(context.Context, *sqlx.Tx) error) error {
	if opts == nil {
		opts = &sql.TxOptions{}
	}
	tx, err := repo.DB.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	for _, fun := range funs {
		if err := fun(ctx, tx); err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
