package dbpostgres

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/server"
)

type DB struct {
	pool *pgxpool.Pool
}

var (
	sqlInsertUserURL = `INSERT INTO users_urls (user_id, url_id) VALUES ($1, $2)
		ON CONFLICT (user_id, url_id) DO NOTHING`
	sqlInsertUrlID = `WITH extant AS (SELECT id FROM urls_id WHERE (url) = ($2)),
						inserted AS (INSERT INTO urls_id (id, url) 
							SELECT ($1), ($2)
							WHERE NOT EXISTS (SELECT NULL FROM extant)
							RETURNING id)
							SELECT id, 'succes' FROM inserted
							UNION ALL
							SELECT id, 'conflict' FROM extant`
)

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

func (db *DB) Add(url string, userID uint64) (string, error) {
	ctx := context.Background()
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	urlID := repositories.ShortenUrlID(url)
	var status string
	var errConflict error

	err = tx.QueryRow(ctx, sqlInsertUrlID, urlID, url).Scan(&urlID, &status)
	if err != nil {
		return "", err
	}
	if status == "conflict" {
		errConflict = app.ErrConflictURLID
	}
	if _, err = tx.Exec(ctx, sqlInsertUserURL, userID, urlID); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return urlID, errConflict
}

func (db *DB) AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error) {
	ctx := context.Background()
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	respBatch := make([]models.ResponseBatch, 0)
	for _, v := range urls {
		urlID := repositories.ShortenUrlID(v.OriginalURL)
		if err = tx.QueryRow(ctx, sqlInsertUrlID, urlID, v.OriginalURL).Scan(&urlID, new(string)); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, sqlInsertUserURL, userID, urlID); err != nil {
			return nil, err
		}
		respBatch = append(respBatch, models.ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      urlID,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return respBatch, nil
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

func (db *DB) GetUserURLs(userID uint64) ([]models.UserURLs, error) {
	urls := make([]models.UserURLs, 0)
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
		var urlID string
		url := &models.UserURLs{}
		err := rows.Scan(&urlID, &url.OriginalURL)
		if err != nil {
			return nil, err
		}
		url.ShortURL = repositories.BaseURL(urlID)
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
