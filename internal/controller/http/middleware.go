package httpctrl

import (
	"github.com/wb-go/wbf/ginext"
)

type logger interface {
	WithFields(keyValues ...interface{}) logger
	Error(err error)
}

// Middleware является прослойкой перед вызываемыми обработчиками.
type Middleware struct {
	log logger
}

// NewMiddleware создает новый Middleware.
func NewMiddleware(log logger) *Middleware {
	return &Middleware{log: log}
}

// ErrHandlingMiddleware логирует ошибки, добавленные через c.Error().
func (m *Middleware) ErrHandlingMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Next()

		for _, err := range c.Errors {
			m.log.Error(err)
		}
	}
}
