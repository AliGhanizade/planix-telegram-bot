package ui

// Lang is the display language of the bot.
type Lang string

// Supported languages. Persian is the default.
const (
	Fa Lang = "fa"
	En Lang = "en"
)

// Normalize maps any value to a supported language.
func Normalize(l string) Lang {
	if l == string(En) {
		return En
	}
	return Fa
}

// Toggle returns the other language.
func (l Lang) Toggle() Lang {
	if l == En {
		return Fa
	}
	return En
}
