package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

func GetAppDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appDataPath := filepath.Join(homeDir, "."+internal.APP_NAME)
	if _, err := os.Stat(appDataPath); os.IsNotExist(err) {
		err := os.Mkdir(appDataPath, 0700)
		if err != nil {
			return "", err
		}
	}
	return appDataPath, nil
}

func WriteJSONFile(path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func WriteConfigJSONFile(data any) error {
	path, err := GetAppDataPath()
	path = filepath.Join(path, internal.CONFIG_FILE_NAME)
	if err != nil {
		return errors.New("Error while writing config: " + err.Error())
	}
	return WriteJSONFile(path, data)
}
