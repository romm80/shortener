package repositories

import (
	"crypto/rand"
	"encoding/base32"
	"testing"

	"github.com/romm80/shortener.git/internal/app/repositories/linkedliststorage"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/service"
)

const triesN = 10000

func BenchmarkAdd(b *testing.B) {

	mapDB, _ := mapstorage.New()
	listDB := linkedliststorage.New()
	var urls []string

	for i := 0; i < triesN; i++ {
		randomBytes := make([]byte, 32)
		_, _ = rand.Read(randomBytes)
		urls = append(urls, base32.StdEncoding.EncodeToString(randomBytes))
	}

	b.ResetTimer()

	b.Run("map", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			_ = mapDB.Add(urls[i], service.ShortenURLID(urls[i]), 1)
		}
	})

	b.Run("list", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			_ = listDB.Add(urls[i], service.ShortenURLID(urls[i]), 1)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	mapDB, _ := mapstorage.New()
	listDB := linkedliststorage.New()
	var urls, IDs []string

	for i := 0; i < triesN; i++ {
		randomBytes := make([]byte, 32)
		_, _ = rand.Read(randomBytes)
		urls = append(urls, base32.StdEncoding.EncodeToString(randomBytes))
	}
	for i := 0; i < triesN; i++ {
		id := service.ShortenURLID(urls[i])
		_ = mapDB.Add(urls[i], id, 1)
		_ = listDB.Add(urls[i], id, 1)
		IDs = append(IDs, id)
	}
	b.ResetTimer()

	b.Run("map", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			_, _ = mapDB.Get(IDs[i])
		}
	})

	b.Run("list", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			_, _ = listDB.Get(IDs[i])
		}
	})
}

func BenchmarkGetUserURLs(b *testing.B) {
	const (
		usersN = 100
		urlsN  = 100
	)
	mapDB, _ := mapstorage.New()
	listDB := linkedliststorage.New()

	for i := 0; i < usersN; i++ {
		for j := 0; j < urlsN; j++ {
			randomBytes := make([]byte, 32)
			_, _ = rand.Read(randomBytes)
			originURL := base32.StdEncoding.EncodeToString(randomBytes)
			id := service.ShortenURLID(originURL)
			_ = mapDB.Add(originURL, id, uint64(i))
			_ = listDB.Add(originURL, id, uint64(i))
		}
	}

	b.ResetTimer()

	b.Run("map", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			_, _ = mapDB.GetUserURLs(uint64(i % usersN))
		}
	})

	b.Run("list", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			_, _ = listDB.GetUserURLs(uint64(i % usersN))
		}
	})

}

func BenchmarkDeleteBatch(b *testing.B) {
	const (
		usersN = 100
		urlsN  = 100
	)
	mapDB, _ := mapstorage.New()
	listDB := linkedliststorage.New()
	userURLs := make(map[uint64][]string, usersN)

	for i := 0; i < usersN; i++ {
		for j := 0; j < urlsN; j++ {
			randomBytes := make([]byte, 32)
			_, _ = rand.Read(randomBytes)
			originURL := base32.StdEncoding.EncodeToString(randomBytes)
			id := service.ShortenURLID(originURL)
			_ = mapDB.Add(originURL, id, uint64(i))
			_ = listDB.Add(originURL, id, uint64(i))
			if userURLs[uint64(i)] == nil {
				userURLs[uint64(i)] = make([]string, 0, urlsN)
			}
			userURLs[uint64(i)] = append(userURLs[uint64(i)], id)
		}
	}

	b.ResetTimer()

	b.Run("map", func(b *testing.B) {
		for i := 0; i < usersN; i++ {
			_ = mapDB.DeleteBatch(uint64(i), userURLs[uint64(i)])
		}
	})

	b.Run("list", func(b *testing.B) {
		for i := 0; i < usersN; i++ {
			_ = listDB.DeleteBatch(uint64(i), userURLs[uint64(i)])
		}
	})
}
