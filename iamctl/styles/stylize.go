package styles

import "github.com/charmbracelet/lipgloss"

func StylizeSuccessMessage(message string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)
	return style.Render("\n[ SUCCESS ]\n" + message + "\n")
}

func StylizeErrorMessage(message string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)
	return style.Render("\n[ ERROR ]\n" + message + "\n")
}

func StylizeInfoMessage(message string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)
	return style.Render("\n[ INFO ]\n" + message + "\n")
}

func StylizeWarningMessage(message string) string {
	boxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FFFF00")).
		Bold(true)
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)
	return boxStyle.Render("[ WARNING ]") + "\n" + messageStyle.Render(message)
}
