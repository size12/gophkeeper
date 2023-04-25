package handlers

import (
	"github.com/size12/gophkeeper/internal/entity"
)

type Client struct {
	Conn      *ClientConn
	authToken entity.AuthToken
}

func NewClient(conn *ClientConn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (client *Client) Login(credentials entity.UserCredentials) error {
	if credentials.Login == "" || credentials.Password == "" {
		return ErrFieldIsEmpty
	}
	authToken, err := client.Conn.Login(credentials)
	if err != nil {
		return err
	}
	client.authToken = entity.AuthToken(authToken)
	return nil
}

func (client *Client) Register(credentials entity.UserCredentials) error {
	authToken, err := client.Conn.Register(credentials)
	if err != nil {
		return err
	}
	client.authToken = entity.AuthToken(authToken)
	return nil
}

func (client *Client) GetRecordsInfo() ([]entity.Record, error) {
	return client.Conn.GetRecordsInfo(client.authToken)
}

func (client *Client) GetRecord(recordID string) (entity.Record, error) {
	// TODO ADD DECODING
	return client.Conn.GetRecord(client.authToken, recordID)
}

func (client *Client) DeleteRecord(recordID string) error {
	return client.Conn.DeleteRecord(client.authToken, recordID)
}

func (client *Client) CreateRecord(record entity.Record) error {
	// TODO ADD ENCODING
	return client.Conn.CreateRecord(client.authToken, record)
}
