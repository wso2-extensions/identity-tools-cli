package utils

import (
	"encoding/base64"
)

func EncodeToBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
