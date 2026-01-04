package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StylizeLoginStatus creates a beautifully formatted login status display
func StylizeLoginStatus(status, orgName, serverName string, err error) string {
	var (
		titleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ff7300")).
				MarginTop(1).
				MarginBottom(1)

		boxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#ff7300")).
				Padding(1, 2).
				MarginTop(1).
				MarginBottom(1).
				Width(60)

		successStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true)

		warningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff7300")).
				Bold(true)

		errorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

		labelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Width(20)

		valueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
	)

	if err != nil {
		return boxStyle.Render(
			titleStyle.Render("Error") + "\n\n" +
				errorStyle.Render(fmt.Sprintf("Error getting login details: %s", err.Error())),
		)
	}

	var content string

	if status == "Logged In" {
		statusLine := successStyle.Render("Logged In")
		orgLine := labelStyle.Render("Organization:") + " " + valueStyle.Render(orgName)
		serverLine := labelStyle.Render("Server:") + " " + valueStyle.Render(serverName)

		content =
			statusLine + "\n\n" +
			orgLine + "\n" +
			serverLine
	} else {
		statusLine := warningStyle.Render("Logged Out")
		message := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff7300")).
			Render("Please login to continue using the tool")

		content =
			statusLine + "\n\n" +
			message
	}

	return boxStyle.Render(content)
}
