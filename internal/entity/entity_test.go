package entity

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordType_String(t *testing.T) {
	tc := []struct {
		name string
		arg  RecordType
		want string
	}{
		{
			"Login and password",
			TypeLoginAndPassword,
			"Login + password",
		},
		{
			"Binary file",
			TypeFile,
			"Binary file",
		},
		{
			"Text",
			TypeText,
			"Text",
		},
		{
			"Credit card",
			TypeCreditCard,
			"Credit card",
		},
		{
			"Unknown message",
			999,
			"Unknown",
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		result := test.arg.String()
		assert.Equal(t, test.want, result)
	}

}

func TestMessage_Bytes(t *testing.T) {
	tc := []struct {
		name string
		arg  interface{ Bytes() ([]byte, error) }
		want []byte
	}{
		{
			"Login and password",
			&LoginAndPassword{
				Login:    "login",
				Password: "password",
			},
			[]byte("login:password"),
		},
		{
			"Text",
			&TextData{
				Text: "hello world!",
			},
			[]byte("hello world!"),
		},
		{
			"Credit card",
			&CreditCard{
				CardNumber:     "2202203293415444",
				ExpirationDate: "08/47",
				CVCCode:        "123",
			},
			[]byte("2202203293415444|08/47|123"),
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		result, err := test.arg.Bytes()
		assert.NoError(t, err)
		assert.Equal(t, test.want, result)
	}
}

func TestBinaryFile_Bytes(t *testing.T) {
	file, err := os.Create("test_file.txt")
	assert.NoError(t, err)
	_, err = file.Write([]byte("hello world!"))
	assert.NoError(t, err)

	record := &BinaryFile{
		FilePath: "test_file.txt",
		File:     file,
	}

	result, err := record.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world!"), result)

	assert.NoError(t, os.RemoveAll("test_file.txt"))
	result, err = record.Bytes()
	assert.Error(t, err)
	assert.Empty(t, result)
}
