package observability

import (
    "os"
    "strings"
    "github.com/rs/zerolog"
)

func NewLogger(level string) *zerolog.Logger {
    lvl := zerolog.InfoLevel
    switch strings.ToLower(level) {
    case "debug":
        lvl = zerolog.DebugLevel
    case "warn":
        lvl = zerolog.WarnLevel
    case "error":
        lvl = zerolog.ErrorLevel
    }
    logger := zerolog.New(os.Stdout).Level(lvl).With().Timestamp().Logger()
    return &logger
}


