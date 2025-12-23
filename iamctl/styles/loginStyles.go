package styles

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	background = lipgloss.AdaptiveColor{Dark: "#000000"} // Black
	selection  = lipgloss.AdaptiveColor{Dark: "#ff7300"} // Orange
	foreground = lipgloss.AdaptiveColor{Dark: "#ffffff"} // White
	comment    = lipgloss.AdaptiveColor{Dark: "#ffffff"} // White
	orange     = lipgloss.AdaptiveColor{Dark: "#ff7300"} // Orange
)

func GetLoginTheme() *huh.Theme {
	var LoginTheme = huh.ThemeBase()
	LoginTheme.Focused.Base = LoginTheme.Focused.Base.BorderForeground(selection)
	LoginTheme.Focused.Card = LoginTheme.Focused.Base
	LoginTheme.Focused.Title = LoginTheme.Focused.Title.Foreground(orange)
	LoginTheme.Focused.NoteTitle = LoginTheme.Focused.NoteTitle.Foreground(orange)
	LoginTheme.Focused.Description = LoginTheme.Focused.Description.Foreground(comment)
	LoginTheme.Focused.ErrorIndicator = LoginTheme.Focused.ErrorIndicator.Foreground(orange)
	LoginTheme.Focused.Directory = LoginTheme.Focused.Directory.Foreground(orange)
	LoginTheme.Focused.File = LoginTheme.Focused.File.Foreground(foreground)
	LoginTheme.Focused.ErrorMessage = LoginTheme.Focused.ErrorMessage.Foreground(orange)
	LoginTheme.Focused.SelectSelector = LoginTheme.Focused.SelectSelector.Foreground(orange)
	LoginTheme.Focused.NextIndicator = LoginTheme.Focused.NextIndicator.Foreground(orange)
	LoginTheme.Focused.PrevIndicator = LoginTheme.Focused.PrevIndicator.Foreground(orange)
	LoginTheme.Focused.Option = LoginTheme.Focused.Option.Foreground(foreground)
	LoginTheme.Focused.MultiSelectSelector = LoginTheme.Focused.MultiSelectSelector.Foreground(orange)
	LoginTheme.Focused.SelectedOption = LoginTheme.Focused.SelectedOption.Foreground(orange)
	LoginTheme.Focused.SelectedPrefix = LoginTheme.Focused.SelectedPrefix.Foreground(orange)
	LoginTheme.Focused.UnselectedOption = LoginTheme.Focused.UnselectedOption.Foreground(foreground)
	LoginTheme.Focused.UnselectedPrefix = LoginTheme.Focused.UnselectedPrefix.Foreground(comment)
	LoginTheme.Focused.FocusedButton = LoginTheme.Focused.FocusedButton.Foreground(orange).Background(orange).Bold(true)
	LoginTheme.Focused.BlurredButton = LoginTheme.Focused.BlurredButton.Foreground(foreground).Background(background)
	LoginTheme.Focused.TextInput.Cursor = LoginTheme.Focused.TextInput.Cursor.Foreground(orange)
	LoginTheme.Focused.TextInput.Placeholder = LoginTheme.Focused.TextInput.Placeholder.Foreground(comment)
	LoginTheme.Focused.TextInput.Prompt = LoginTheme.Focused.TextInput.Prompt.Foreground(orange)

	LoginTheme.Blurred = LoginTheme.Focused
	LoginTheme.Blurred.Base = LoginTheme.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	LoginTheme.Blurred.Card = LoginTheme.Blurred.Base
	LoginTheme.Blurred.NextIndicator = lipgloss.NewStyle()
	LoginTheme.Blurred.PrevIndicator = lipgloss.NewStyle()

	LoginTheme.Group.Title = LoginTheme.Focused.Title
	LoginTheme.Group.Description = LoginTheme.Focused.Description

	return LoginTheme
}

func GetSpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(orange).Bold(true)
}
