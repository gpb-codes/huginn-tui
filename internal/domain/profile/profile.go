package profile

import "time"

// Profile represents the persistent user profile (Markdown backed).
// Stored at ~/.huginn/profile.md
type Profile struct {
	Version       int           `json:"version"`
	Communication Communication `json:"communication"`
	Development   Development   `json:"development"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type Communication struct {
	Language       string `json:"language"`
	Style          string `json:"style"`
	Verbosity      string `json:"verbosity"`
	TechnicalDepth string `json:"technical_depth"`
}

type Development struct {
	PreferredLanguages []string `json:"preferred_languages"`
	PreferredEditor    string   `json:"preferred_editor"`
}

func Default() Profile {
	return Profile{
		Version: 1,
		Communication: Communication{
			Language:       "es",
			Style:          "direct",
			Verbosity:      "medium",
			TechnicalDepth: "high",
		},
		Development: Development{
			PreferredLanguages: []string{"TypeScript", "Go", "Python"},
			PreferredEditor:    "VS Code",
		},
		UpdatedAt: time.Now(),
	}
}
