package utils

import (
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
	"github.com/zalando/go-keyring"
)

func StoretoKeyring(key string, value string) error {
	err := keyring.Set(internal.APP_NAME, key, value)
	return err
}

func GetfromKeyring(key string) (string, error) {
	value, err := keyring.Get(internal.APP_NAME, key)
	return value, err
}

func DeletefromKeyring(key string) error {
	err := keyring.Delete(internal.APP_NAME, key)
	return err
}
