package domain

// ColorMode управляет раскраской вывода.
type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

// Options — параметры CLI‑вывода.
type Options struct {
	Preset           string
	Fields           map[string]bool // включённые секции
	BodyPreviewBytes int
	Color            ColorMode
	Filter           string // простая подстрока для фильтрации
}

// EnsureFieldsFromPreset заполняет Fields на основе пресета, если явно не указано.
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
		// по умолчанию basic
		o.Fields["line"] = true
		o.Fields["sizes"] = true
	}
}
