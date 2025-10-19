package logger

import (
	"github.com/wb-go/wbf/zlog"
)

type loggerAdapter struct {
	impl zlog.Zerolog
}

func (a loggerAdapter) WithFields(keyValues ...interface{}) Logger {
	return loggerAdapter{impl: a.impl.With().Fields(keyValues).Logger()}
}

func (a loggerAdapter) Error(err error) {
	a.impl.Err(err).Send()
}

func (a loggerAdapter) Debug(msg string) {
	a.impl.Debug().Msg(msg)
}

// NewLoggerAdapter создает новый адаптер для zlog, имплементирующий интерфейс logger.
func NewLoggerAdapter(z zlog.Zerolog) Logger {
	return loggerAdapter{impl: z}
}
