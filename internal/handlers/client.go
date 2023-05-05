package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"os"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage"
)

// Client struct for client handlers.
type Client struct {
	Conn      ClientConn
	authToken entity.AuthToken
	masterKey []byte
}

// NewClientHandlers returns new client handlers.
func NewClientHandlers(conn ClientConn) *Client {
	return &Client{
		Conn: conn,
	}
}

// Login logins user by login and password.
func (client *Client) Login(credentials entity.UserCredentials) error {
	if credentials.Login == "" || credentials.Password == "" || len(credentials.MasterKey) == 0 {
		return ErrFieldIsEmpty
	}
	authToken, err := client.Conn.Login(credentials)
	if err != nil {
		return err
	}
	client.authToken = entity.AuthToken(authToken)
	sha := sha256.New()

	_, err = sha.Write(client.masterKey)

	if err != nil {
		return storage.ErrUnknown
	}

	key := sha.Sum(nil)

	client.masterKey = key
	return nil
}

// Register creates new user by login and password.
func (client *Client) Register(credentials entity.UserCredentials) error {
	if credentials.Login == "" || credentials.Password == "" || len(credentials.MasterKey) == 0 {
		return ErrFieldIsEmpty
	}
	authToken, err := client.Conn.Register(credentials)
	if err != nil {
		return err
	}
	client.authToken = entity.AuthToken(authToken)

	sha := sha256.New()

	_, err = sha.Write(client.masterKey)

	if err != nil {
		return storage.ErrUnknown
	}

	key := sha.Sum(nil)

	client.masterKey = key
	return nil
}

// GetRecordsInfo gets all records.
func (client *Client) GetRecordsInfo() ([]entity.Record, error) {
	return client.Conn.GetRecordsInfo(client.authToken)
}

// GetRecord gets record by recordID and decodes it.
func (client *Client) GetRecord(recordID string) (entity.Record, error) {

	record, err := client.Conn.GetRecord(client.authToken, recordID)

	if err != nil {
		return record, err
	}

	aesblock, err := aes.NewCipher(client.masterKey)

	if err != nil {
		return record, ErrWrongMasterKey
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return record, storage.ErrUnknown
	}

	nonce := record.Data[:aesgcm.NonceSize()]

	decoded, err := aesgcm.Open(nil, nonce, record.Data[aesgcm.NonceSize():], nil)

	if err != nil {
		return record, storage.ErrUnknown
	}

	record.Data = decoded

	if record.Type == entity.TypeFile {
		file, err := os.Create(record.Metadata)
		if err != nil {
			return record, storage.ErrUnknown
		}

		_, err = file.Write(record.Data)
		if err != nil {
			return record, storage.ErrUnknown
		}
		record.Data = []byte("Saved file successfully to " + record.Metadata + ".")
	}

	return record, nil
}

// DeleteRecord deletes record by his ID.
func (client *Client) DeleteRecord(recordID string) error {
	return client.Conn.DeleteRecord(client.authToken, recordID)
}

// CreateRecord creates new record.
func (client *Client) CreateRecord(record entity.Record) error {
	aesblock, err := aes.NewCipher(client.masterKey)

	if err != nil {
		return ErrWrongMasterKey
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return storage.ErrUnknown
	}

	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return storage.ErrUnknown
	}

	out := aesgcm.Seal(nil, nonce, record.Data, nil) // зашифровываем

	record.Data = append(nonce, out...)

	return client.Conn.CreateRecord(client.authToken, record)
}

// generateRandom generates random bytes for encrypting.
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
