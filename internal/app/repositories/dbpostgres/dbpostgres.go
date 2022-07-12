// Package dbpostgres implements work with the postgres
package dbpostgres

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
)

type DB struct {
	pool *pgxpool.Pool
}

var (
	sqlInsertURLID = `WITH 	extant AS 	(SELECT url_id FROM urls_id WHERE url = ($2)),
							inserted AS (INSERT INTO urls_id (url_id, url, user_id) SELECT ($1), ($2), ($3)
											WHERE NOT EXISTS (SELECT NULL FROM extant)
											RETURNING url_id)
							SELECT url_id, 'succes' FROM inserted
							UNION ALL
							SELECT url_id, 'conflict' FROM extant`
)

func New() (*DB, error) {

	if err := migrateDB(); err != nil {
		return nil, err
	}

	pool, err := pgxpool.Connect(context.Background(), app.Cfg.DatabaseDNS)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func migrateDB() error {

	m, err := migrate.New(
		"file://db/migrations",
		app.Cfg.DatabaseDNS)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err.Error() != "no change" {
		return err
	}
	return nil
}

func (db *DB) Add(url, urlID string, userID uint64) error {
	ctx := context.Background()
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var errConflict error

	err = tx.QueryRow(ctx, sqlInsertURLID, urlID, url, userID).Scan(&urlID, &status)
	if err != nil {
		return err
	}
	if status == "conflict" {
		errConflict = app.ErrConflictURLID
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return errConflict
}

func (db *DB) Get(id string) (originURL string, err error) {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()

	deleted := false
	err = conn.QueryRow(context.Background(), `SELECT url, deleted FROM urls_id WHERE url_id=$1`, id).Scan(&originURL, &deleted)
	if err != nil {
		return
	}
	if deleted {
		err = app.ErrDeletedURL
	}
	return
}

func (db *DB) GetUserURLs(userID uint64) ([]models.UserURLs, error) {
	urls := make([]models.UserURLs, 0)
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return urls, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `SELECT url_id, url FROM urls_id WHERE user_id=($1)`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var urlID string
		url := &models.UserURLs{}
		err := rows.Scan(&urlID, &url.OriginalURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, *url)
	}
	return urls, nil
}

func (db *DB) NewUser() (userID uint64, err error) {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()

	err = conn.QueryRow(context.Background(), `INSERT INTO users (id) VALUES(default) RETURNING (id)`).Scan(&userID)
	if err != nil {
		return
	}
	return
}

func (db *DB) Ping() error {
	return db.pool.Ping(context.Background())
}

func (db *DB) DeleteBatch(userID uint64, urlsID []string) error {

	ctx := context.Background()
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE urls_id SET deleted=true WHERE user_id = ($1) AND url_id = any($2)`, userID, urlsID)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (db *DB) GetStats() (*models.StatsResponse, error) {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	res := models.StatsResponse{}
	q := `SELECT SUM(urls) AS urls, SUM(users) AS users FROM 
			(SELECT COUNT(*) AS urls, 0 AS users FROM urls_id WHERE NOT deleted
			UNION ALL
			SELECT 0, COUNT(*) FROM users) AS res`
	err = conn.QueryRow(context.Background(), q).Scan(&res.URLs, &res.Users)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
