package utils

import (
	//"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/components"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

func EnsureAppDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appDataPath := filepath.Join(homeDir, "_"+internal.APP_NAME)
	if _, err := os.Stat(appDataPath); os.IsNotExist(err) {
		err := os.Mkdir(appDataPath, 0700)
		if err != nil {
			return "", err
		}
	}
	return appDataPath, nil
}

func GetAppDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	appDataPath := filepath.Join(homeDir, "_"+internal.APP_NAME)
	if _, err := os.Stat(appDataPath); os.IsNotExist(err) {
		return "", errors.New("app data directory does not exist")
	}
	return appDataPath, nil
}

// func WriteJSONFile(path string, data any) error {
// 	file, err := os.Create(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "  ")
// 	return encoder.Encode(data)
// }

// func WriteConfigJSONFile(data any) error {
// 	path, err := GetAppDataPath()
// 	path = filepath.Join(path, internal.CONFIG_FILE_NAME)
// 	if err != nil {
// 		return errors.New("Error while writing config: " + err.Error())
// 	}
// 	return WriteJSONFile(path, data)
// }

// func DeleteConfigJSONFile() error {
// 	path, err := GetAppDataPath()
// 	path = filepath.Join(path, internal.CONFIG_FILE_NAME)
// 	if err != nil {
// 		return errors.New("Error while deleting config: " + err.Error())
// 	}
// 	return os.Remove(path)
// }

func CreateConfigFileIfNotExists(dirPath string) error {
	filePath := filepath.Join(dirPath, internal.CONFIG_FILE_NAME+"."+internal.CONFIG_FILE_TYPE)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		if internal.CONFIG_FILE_TYPE == "json" {
			_, err = file.WriteString("{}")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func InitializeViperConfig() error {
	dirPath, err := EnsureAppDataPath()
	if err != nil {
		return errors.New("Error while initializing config: " + err.Error())
	}

	fullPath := filepath.Join(dirPath, internal.CONFIG_FILE_NAME+"."+internal.CONFIG_FILE_TYPE)
	viper.SetConfigFile(fullPath)
	viper.SetConfigType(internal.CONFIG_FILE_TYPE)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) || os.IsNotExist(err) {
			if err := CreateConfigFileIfNotExists(dirPath); err != nil {
				return errors.New("Failed to create config file: " + err.Error())
			}
			log.Printf(components.StylizeInfoMessage("Created config file at %s"), fullPath)

			// Try reading again now that the file exists and has valid content
			if err := viper.ReadInConfig(); err != nil {
				return errors.New("created config file but failed to read it: " + err.Error())
			}
		} else {
			// If the error was NOT "File Not Found" (e.g., permission denied), fail immediately
			return errors.New("Config file found but unreadable: " + err.Error())
		}
	}
	return nil
}

func GetConfigValue(key string) (string, error) {
	value := viper.GetString(key)
	if value == "" {
		return "", errors.New("Config key not found: " + key)
	}
	return value, nil
}

func SetConfigValue(key string, value string) error {
	viper.Set(key, value)
	err := viper.WriteConfig()
	return err
}

func DeleteConfigValue(key string) error {
	viper.Set(key, nil)
	err := viper.WriteConfig()
	return err
}

func ClearConfigOnLogout() error {
	err := DeleteConfigValue(internal.ORG_NAME_KEY)
	if err != nil {
		return err
	}
	err = DeleteConfigValue(internal.CLIENT_ID_KEY)
	if err != nil {
		return err
	}
	err = DeleteConfigValue(internal.SERVER_URL_KEY)
	if err != nil {
		return err
	}
	return nil
}
