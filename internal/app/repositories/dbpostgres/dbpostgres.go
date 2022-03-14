package dbpostgres

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/server"
)

type DB struct {
	pool *pgxpool.Pool
}

func New() (*DB, error) {

	if err := migrateDB(); err != nil {
		return nil, err
	}

	pool, err := pgxpool.Connect(context.Background(), server.Cfg.DatabaseDNS)
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
		server.Cfg.DatabaseDNS)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err.Error() != "no change" {
		return err
	}
	return nil
}

func (db *DB) Add(urlID, url string, userID uint64) error {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `INSERT INTO urls_id (id, url) VALUES($1, $2) 
	ON CONFLICT (id) DO UPDATE SET url=excluded.url`, urlID, url)
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), `INSERT INTO users_urls (user_id, url_id) VALUES ($1, $2)
	ON CONFLICT (user_id, url_id) DO NOTHING`, userID, urlID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Get(id string) (originURL string, err error) {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()

	err = conn.QueryRow(context.Background(), "SELECT url FROM urls_id WHERE id=$1", id).Scan(&originURL)
	if err != nil {
		return
	}
	return
}

func (db *DB) GetUserURLs(userID uint64) ([]repositories.UserURLs, error) {
	urls := make([]repositories.UserURLs, 0)
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return urls, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `SELECT t1.url_id, t2.url FROM users_urls AS t1 
	INNER JOIN urls_id AS t2 ON t1.url_id = t2.id AND t1.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		url := &repositories.UserURLs{}
		err := rows.Scan(&url.ShortURL, &url.OriginalURL)
		if err != nil {
			return nil, err
		}
		url.ShortURL = fmt.Sprintf("%s/%s", server.Cfg.BaseURL, url.ShortURL)
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

func (db *DB) CheckUserID(userID uint64) (bool, error) {
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Release()

	row, err := conn.Query(context.Background(), `SELECT id FROM users WHERE id=$1`, userID)
	if err != nil {
		return false, err
	}
	if row.Next() {
		return true, nil
	}
	return false, nil
}

func (db *DB) Ping() error {
	return db.pool.Ping(context.Background())
}
