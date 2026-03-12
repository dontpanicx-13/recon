package theme

import (
	"encoding/json"
	"os"
)

type Theme struct {
	AppBg         string `json:"app_bg"`
	AccentBg      string `json:"accent_bg"`
	AccentFg      string `json:"accent_fg"`
	StatusBg      string `json:"status_bg"`
	StatusFg      string `json:"status_fg"`
	SpinnerFg     string `json:"spinner_fg"`
	ControlsFg    string `json:"controls_fg"`
}

func Default() Theme {
	return Theme{
		AppBg:         "#1e2030",
		AccentBg:      "#DFC2FC",
		AccentFg:      "#1e2030",
		StatusBg:      "#2b2f45",
		StatusFg:      "#e6e6e6",
		SpinnerFg:     "#99e174",
		ControlsFg:    "#82aaff",
	}
}

func Load() Theme {
	theme := Default()
	data, err := os.ReadFile("ui_colors.json")
	if err != nil {
		return theme
	}

	var override Theme
	if err := json.Unmarshal(data, &override); err != nil {
		return theme
	}

	return theme.merge(override)
}

func (t Theme) merge(o Theme) Theme {
	if o.AppBg != "" {
		t.AppBg = o.AppBg
	}
	if o.AccentBg != "" {
		t.AccentBg = o.AccentBg
	}
	if o.AccentFg != "" {
		t.AccentFg = o.AccentFg
	}
	if o.StatusBg != "" {
		t.StatusBg = o.StatusBg
	}
	if o.StatusFg != "" {
		t.StatusFg = o.StatusFg
	}
	if o.SpinnerFg != "" {
		t.SpinnerFg = o.SpinnerFg
	}
	if o.ControlsFg != "" {
		t.ControlsFg = o.ControlsFg
	}

	return t
}
