package utils

import (
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
	"github.com/zalando/go-keyring"
)

type KeyRingProvider interface {
	Set(service, key, value string) error
	Get(service, key string) (string, error)
	Delete(service, key string) error
}

type DefaultKeyRing struct{}

func (d DefaultKeyRing) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}
func (d DefaultKeyRing) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}
func (d DefaultKeyRing) Delete(service, key string) error {
	return keyring.Delete(service, key)
}

var keyringStore KeyRingProvider = &DefaultKeyRing{}

func StoretoKeyring(key string, value string) error {
	err := keyringStore.Set(internal.APP_NAME, key, value)
	return err
}

func GetfromKeyring(key string) (string, error) {
	value, err := keyringStore.Get(internal.APP_NAME, key)
	return value, err
}

func DeletefromKeyring(key string) error {
	err := keyringStore.Delete(internal.APP_NAME, key)
	return err
}
