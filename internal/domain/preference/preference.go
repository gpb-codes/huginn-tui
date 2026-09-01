package preference

import "time"

// Preference is a learned user preference, stored as Markdown.
type Preference struct {
	ID         string    `json:"id"`
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Importance float64   `json:"importance"`
	Confidence float64   `json:"confidence"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Source     string    `json:"source"`
}

func New(id, key, value string, importance, confidence float64) Preference {
	now := time.Now()
	return Preference{
		ID:         id,
		Key:        key,
		Value:      value,
		Importance: importance,
		Confidence: confidence,
		CreatedAt:  now,
		UpdatedAt:  now,
		Source:     "user",
	}
}
