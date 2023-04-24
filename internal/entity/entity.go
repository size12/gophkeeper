package entity

import (
	"io"
	"os"
)

type UserCredentials struct {
	Login    string
	Password string
}

type UserID string
type AuthToken string

type Record struct {
	ID       string
	Metadata string
	Type     string
	Data     []byte
}

type LoginAndPassword struct {
	Login    string
	Password string
}

func (data *LoginAndPassword) Bytes() ([]byte, error) {
	return []byte(data.Login + ":" + data.Password), nil
}

type TextData struct {
	Text string
}

func (data *TextData) Bytes() ([]byte, error) {
	return []byte(data.Text), nil
}

type BinaryFile struct {
	File *os.File
}

func (data *BinaryFile) Bytes() ([]byte, error) {
	return io.ReadAll(data.File)
}

type CreditCard struct {
	CardNumber     string
	ExpirationDate string
	CVCCode        string
}

func (data *CreditCard) Bytes() ([]byte, error) {
	return []byte(data.CardNumber + "|" + data.ExpirationDate + "|" + data.CVCCode), nil
}
