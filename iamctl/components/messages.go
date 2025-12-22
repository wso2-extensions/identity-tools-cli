package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Use slightly off-white (#FAFAFA) to prevent terminals from
	// "smart inverting" pure white to black.
	whiteText = lipgloss.Color("#FAFAFA")

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteText).
			Background(lipgloss.Color("#049409"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteText).
			Background(lipgloss.Color("#FF0000"))

	InfoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteText).
			Background(lipgloss.Color("#0099FF"))

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteText).
			Background(lipgloss.Color("#D3BB07"))
)

func StylizeSuccessMessage(message string) string {
	// 2. Render the block HERE, inside the function
	successBlock := "\n" + SuccessStyle.Render("SUCCESS") + "\n"

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	// Note: Removed 'components.' prefix since we are inside the package
	return fmt.Sprintf("%s \n %s", successBlock, messageStyle.Render(message+"\n"))
}

func StylizeErrorMessage(message string) string {
	errorBlock := "\n" + ErrorStyle.Render("ERROR") + "\n"
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)
	return fmt.Sprintf("%s \n %s", errorBlock, messageStyle.Render(message+"\n"))
}

func StylizeInfoMessage(message string) string {
	infoBlock := "\n" + InfoStyle.Render("INFO") + "\n"
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)
	return fmt.Sprintf("%s \n %s", infoBlock, messageStyle.Render(message+"\n"))
}

func StylizeWarningMessage(message string) string {
	warningBlock := "\n" + WarningStyle.Render("WARNING") + "\n"
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)
	return fmt.Sprintf("%s \n %s", warningBlock, messageStyle.Render(message+"\n"))
}
