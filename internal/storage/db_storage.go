package storage

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/size12/gophkeeper/internal/entity"
)

// DBStorage for db storage.
type DBStorage struct {
	DB *sql.DB
}

// NewDBStorage connects to DB.
func NewDBStorage(connectionURL string) *DBStorage {
	db, err := sql.Open("pgx", connectionURL)

	if err != nil {
		log.Fatalln("Failed open DB storage:", err)
		return nil
	}

	return &DBStorage{DB: db}
}

// MigrateUP migrates DB.
func (storage *DBStorage) MigrateUP() {
	driver, err := postgres.WithInstance(storage.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed create postgres instance: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)

	if err != nil {
		log.Fatalf("Failed create migration instance: %v\n", err)
		return
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalln("Failed migrate: ", err)
		return
	}
}

// CreateUser saves to DB new user.
func (storage *DBStorage) CreateUser(credentials entity.UserCredentials) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := storage.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE login = $1`, credentials.Login)

	sameLoginCounter := 0
	err := row.Scan(&sameLoginCounter)

	if err != nil || row.Err() != nil {
		log.Println("Failed get row while checking for login conflict:", err)
		return ErrUnknown
	}

	if sameLoginCounter > 0 {
		return ErrLoginExists
	}

	_, err = storage.DB.ExecContext(ctx, `INSERT INTO users (login, password) VALUES ($1, $2)`, credentials.Login, credentials.Password)
	if err != nil {
		log.Println("Failed insert new user into table users:", err)
		return ErrUnknown
	}

	return nil
}

// LoginUser check if credentials are valid. Returns userID.
func (storage *DBStorage) LoginUser(credentials entity.UserCredentials) (entity.UserID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := storage.DB.QueryRowContext(ctx, `SELECT user_id FROM users WHERE login = $1 AND password = $2`, credentials.Login, credentials.Password)

	var userID entity.UserID
	err := row.Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		return userID, ErrWrongCredentials
	}

	if err != nil || row.Err() != nil {
		log.Println("Failed get row while checking for correct user credentials:", err)
		return userID, ErrUnknown
	}

	return userID, nil
}

// GetRecordsInfo gets all DB record from this user.
func (storage *DBStorage) GetRecordsInfo(ctx context.Context) ([]entity.Record, error) {
	userID, ok := ctx.Value("userID").(entity.UserID)
	if !ok {
		log.Println("Failed get userID from context in getting all records")
		return nil, ErrUserUnauthorized
	}

	rows, err := storage.DB.QueryContext(ctx, `SELECT record_id, record_type, metadata FROM users_data WHERE user_id = $1`, userID)
	if err != nil {
		log.Println("Failed get rows in getting all records:", err)
		return nil, ErrUnknown
	}

	defer rows.Close()

	result := make([]entity.Record, 0, 10)
	var row entity.Record
	for rows.Next() {
		err := rows.Scan(&row.ID, &row.Type, &row.Metadata)

		if err != nil {
			log.Println("Failed get next row in getting all records:", err)
			return nil, ErrUnknown
		}

		result = append(result, row)
	}

	if rows.Err() != nil {
		log.Println("Failed get rows in getting all records:", err)
		return nil, ErrUnknown
	}

	return result, nil
}

// CreateRecord saves new record to DB, returns recordID.
func (storage *DBStorage) CreateRecord(ctx context.Context, record entity.Record) (string, error) {
	userID, ok := ctx.Value("userID").(entity.UserID)
	if !ok {
		log.Println("Failed get userID from context in getting all records")
		return "", ErrUserUnauthorized
	}

	hexDataString := hex.EncodeToString(record.Data)

	row := storage.DB.QueryRowContext(ctx, `INSERT INTO users_data (user_id, record_type, metadata, encoded_data) VALUES ($1, $2, $3, $4) RETURNING record_id`, userID, record.Type, record.Metadata, hexDataString)

	recordID := ""

	err := row.Scan(&recordID)
	if err != nil || row.Err() != nil {
		return "", ErrUnknown
	}

	return recordID, nil
}

// GetRecord gets record from DB by ID.
func (storage *DBStorage) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
	record := entity.Record{}

	userID, ok := ctx.Value("userID").(entity.UserID)
	if !ok {
		log.Println("Failed get userID from context in getting all records")
		return record, ErrUserUnauthorized
	}

	row := storage.DB.QueryRowContext(ctx, `SELECT record_id, record_type, metadata, encoded_data FROM users_data WHERE record_id = $1 AND user_id = $2`, recordID, userID)

	hexDataString := ""

	err := row.Scan(&record.ID, &record.Type, &record.Metadata, &hexDataString)

	if errors.Is(err, sql.ErrNoRows) {
		return record, ErrNotFound
	}

	if err != nil || row.Err() != nil {
		log.Println("Failed scan rows to find needed record.", err)
		return record, ErrUnknown
	}

	record.Data, err = hex.DecodeString(hexDataString)

	if err != nil {
		log.Println("Failed convert record data from hex to bytes:", err)
		return record, ErrUnknown
	}

	return record, nil
}

// DeleteRecord deletes record from DB by ID.
func (storage *DBStorage) DeleteRecord(ctx context.Context, recordID string) error {
	userID, ok := ctx.Value("userID").(entity.UserID)
	if !ok {
		log.Println("Failed get userID from context in getting all records")
		return ErrUserUnauthorized
	}

	result, err := storage.DB.ExecContext(ctx, `DELETE FROM users_data WHERE record_id = $1 AND user_id = $2`, recordID, userID)

	if err != nil {
		log.Println("Failed delete record:", err)
		return ErrUnknown
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Failed get affected records:", err)
		return ErrUnknown
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
