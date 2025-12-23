package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockKeyring struct {
	store map[string]string
}

func (m *mockKeyring) Set(service, key, value string) error {
	if m.store == nil {
		m.store = make(map[string]string)
	}
	m.store[key] = value
	return nil
}

func (m *mockKeyring) Get(service, key string) (string, error) {
	val, ok := m.store[key]
	if !ok {
		// Return a standard error or the one expected by your internal logic
		return "", errors.New("secret not found in keyring")
	}
	return val, nil
}

func (m *mockKeyring) Delete(service, key string) error {
	delete(m.store, key)
	return nil
}

func TestStoretoKeyring(t *testing.T) {

	mock := &mockKeyring{
		store: make(map[string]string),
	}

	originalKeyring := keyringStore
	keyringStore = mock
	defer func() { keyringStore = originalKeyring }()

	t.Run("Should successfully store value", func(t *testing.T) {
		key := "test_client_id"
		value := "12345-abcde"

		err := StoretoKeyring(key, value)

		assert.NoError(t, err)

		assert.Equal(t, value, mock.store[key])
	})
}

func TestGetfromKeyring(t *testing.T) {
	// 1. Pre-load the mock with data
	mockData := map[string]string{
		"existing_key": "secret_value",
	}
	mock := &mockKeyring{store: mockData}

	keyringStore = mock
	originalKeyring := keyringStore
	defer func() { keyringStore = originalKeyring }()
	t.Run("Should retrieve existing value", func(t *testing.T) {
		val, err := GetfromKeyring("existing_key")

		assert.NoError(t, err)
		assert.Equal(t, "secret_value", val)
	})

	t.Run("Should return error for missing key", func(t *testing.T) {
		_, err := GetfromKeyring("non_existent_key")

		assert.Error(t, err)
	})
}

func TestDeletefromKeyring(t *testing.T) {
	mock := &mockKeyring{
		store: map[string]string{
			"key_to_delete": "value",
		},
	}
	originalKeyring := keyringStore
	keyringStore = mock
	defer func() { keyringStore = originalKeyring }()

	t.Run("Should delete existing value", func(t *testing.T) {
		err := DeletefromKeyring("existing_key")

		assert.NoError(t, err)
	})

	t.Run("Should return error for missing key", func(t *testing.T) {
		err := DeletefromKeyring("non_existent_key")

		assert.Error(t, err)
	})
}
