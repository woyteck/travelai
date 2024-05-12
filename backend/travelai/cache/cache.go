package cache

import (
	"database/sql"
	"time"
)

type Cache struct {
	db *sql.DB
}

func New(db *sql.DB) Cache {
	return Cache{
		db: db,
	}
}

func (c *Cache) Get(key string) string {
	cacheValue := ""
	err := c.db.QueryRow("SELECT cache_value FROM cache WHERE cache_key=$1 AND valid_until>$2", key, time.Now()).Scan(&cacheValue)
	if err != nil {
		return ""
	}

	return cacheValue
}

func (c *Cache) Set(key string, value string, validityDuration time.Duration) error {
	validUntil := time.Now().Add(validityDuration)
	_, err := c.db.Exec("INSERT INTO cache (created_at, valid_until, cache_key, cache_value) VALUES ($1, $2, $3, $4)", time.Now(), validUntil, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) ColelctGarbage() error {
	_, err := c.db.Exec("DELETE FROM cache WHERE valid_until<$1", time.Now())
	if err != nil {
		return err
	}

	return nil
}
