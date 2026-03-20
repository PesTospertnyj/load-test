package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"load-test/internal/config"
	"load-test/internal/database"
	"load-test/internal/models"
	"load-test/internal/repository"
	"load-test/pkg/postgres"

	"github.com/go-faker/faker/v4"
	log "github.com/sirupsen/logrus"
)

type FakeBook struct {
	Title  string `faker:"sentence"`
	Author string `faker:"name"`
}

var bookGenres = []string{
	"Programming", "Science Fiction", "Fantasy", "Mystery", "Thriller",
	"Romance", "Biography", "History", "Self-Help", "Business",
}

func generateISBN() string {
	return fmt.Sprintf("978-%d-%d-%d-%d",
		rand.Intn(10),
		rand.Intn(100000),
		rand.Intn(1000),
		rand.Intn(10))
}

func generateBookTitle(genre string) string {
	templates := []string{
		"The Art of %s",
		"Mastering %s",
		"Introduction to %s",
		"Advanced %s Techniques",
		"The Complete Guide to %s",
		"%s for Beginners",
		"Professional %s",
		"Modern %s Practices",
		"The %s Handbook",
		"Essential %s",
	}

	template := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(template, genre)
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Info("Starting database seeding...")

	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to PostgreSQL")
	}
	defer pool.Close()

	migrationsPath := "file://db/migrations"
	if err := database.RunMigrations(cfg.Database.DSN(), migrationsPath); err != nil {
		log.WithError(err).Fatal("Failed to run migrations")
	}

	repo := repository.NewPostgresBookRepository(pool)
	log.Info("PostgreSQL repository initialized successfully")

	existingBooks, err := repo.GetAll()
	if err != nil {
		log.WithError(err).Fatal("Failed to check existing books")
	}

	if len(existingBooks) > 0 {
		log.WithField("count", len(existingBooks)).Warn("Database already contains books")
		log.Info("Do you want to continue adding more books? (This will add 10 more)")
	}

	log.Info("Generating 10,000 fake book records...")

	for i := 0; i < 10_000; i++ {
		fakeBook := FakeBook{}
		err := faker.FakeData(&fakeBook)
		if err != nil {
			log.WithError(err).Error("Failed to generate fake data")
			continue
		}

		genre := bookGenres[rand.Intn(len(bookGenres))]
		title := generateBookTitle(genre)

		book := &models.Book{
			Title:  title,
			Author: fakeBook.Author,
			ISBN:   generateISBN(),
		}

		err = repo.Create(book)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"title":  book.Title,
				"author": book.Author,
				"isbn":   book.ISBN,
			}).Error("Failed to create book")
			continue
		}

		log.WithFields(log.Fields{
			"id":     book.ID,
			"title":  book.Title,
			"author": book.Author,
			"isbn":   book.ISBN,
		}).Info("Book created")
	}

	allBooks, err := repo.GetAll()
	if err != nil {
		log.WithError(err).Fatal("Failed to retrieve books")
	}

	log.WithField("total", len(allBooks)).Info("Database seeding completed successfully")

	log.Info("\nBooks in database:")
	for _, book := range allBooks {
		fmt.Printf("  [%d] %s by %s (ISBN: %s)\n", book.ID, book.Title, book.Author, book.ISBN)
	}
}
