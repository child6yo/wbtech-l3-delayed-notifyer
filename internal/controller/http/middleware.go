package httpctrl

import (
	"github.com/child6yo/wbtech-l3-delayed-notifyer/internal/infrastructure/logger"
	"github.com/wb-go/wbf/ginext"
)

// Middleware является прослойкой перед вызываемыми обработчиками.
type Middleware struct {
	logger logger.Logger
}

// NewMiddleware создает новый Middleware.
func NewMiddleware(logger logger.Logger) *Middleware {
	return &Middleware{logger: logger}
}

// ErrHandlingMiddleware логирует ошибки, добавленные через c.Error().
func (m *Middleware) ErrHandlingMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Next()

		metadata := []interface{}{}

		if r, ok := c.Get("request"); ok {
			metadata = append(metadata, []interface{}{"request", r}...)
		}

		for _, err := range c.Errors {
			m.logger.WithFields(metadata...).Error(err)
		}

		c.Errors = nil
	}
}
