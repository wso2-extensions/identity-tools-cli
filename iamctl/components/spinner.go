package components

import (
	"github.com/charmbracelet/huh/spinner"
)

func GetSpinner(text string) *spinner.Spinner {
	return spinner.New().Title(text)
}
