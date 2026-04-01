package output

import "encoding/json"

// Status represents the outcome of a diagnostic check.
type Status int

const (
	Success Status = iota
	Warning
	Error
)

func (s Status) String() string {
	switch s {
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// Result is a single diagnostic line item with optional remediation hint.
type Result struct {
	Category   string `json:"category"`
	Check      string `json:"check"`
	Status     Status `json:"-"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Detail     string `json:"detail,omitempty"`
	// SortOrder breaks ties when ordering results within the same category (lower first). Omitted from JSON.
	SortOrder int `json:"-"`
}

type resultJSON struct {
	Category   string `json:"category"`
	Check      string `json:"check"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

// MarshalJSON encodes status as a string for stable API output.
func (r Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(resultJSON{
		Category:   r.Category,
		Check:      r.Check,
		Status:     r.Status.String(),
		Message:    r.Message,
		Suggestion: r.Suggestion,
		Detail:     r.Detail,
	})
}
