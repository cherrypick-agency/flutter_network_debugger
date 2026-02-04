package domain

// ColorMode controls output coloring.
type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

// Options — CLI output parameters.
type Options struct {
	Preset           string
	Fields           map[string]bool // enabled sections
	BodyPreviewBytes int
	Color            ColorMode
	Filter           string // simple substring for filtering
}

// EnsureFieldsFromPreset populates Fields based on preset, if not explicitly specified.
func (o *Options) EnsureFieldsFromPreset() {
	if len(o.Fields) > 0 {
		return
	}
	o.Fields = map[string]bool{}
	switch o.Preset {
	case "minimal":
		o.Fields["line"] = true
	case "basic":
		o.Fields["line"] = true
		o.Fields["sizes"] = true
	case "advanced":
		o.Fields["line"] = true
		o.Fields["sizes"] = true
		o.Fields["timings"] = true
		o.Fields["req-headers"] = true
		o.Fields["resp-headers"] = true
	case "full":
		o.Fields["line"] = true
		o.Fields["sizes"] = true
		o.Fields["timings"] = true
		o.Fields["req-headers"] = true
		o.Fields["resp-headers"] = true
		o.Fields["req-body"] = true
		o.Fields["resp-body"] = true
		o.Fields["tls"] = true
		o.Fields["cookies"] = true
		o.Fields["ids"] = true
	default:
		// default is basic
		o.Fields["line"] = true
		o.Fields["sizes"] = true
	}
}
