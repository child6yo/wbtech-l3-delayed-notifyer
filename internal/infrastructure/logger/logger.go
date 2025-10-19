package logger

// Logger определяет общий для сервиса интерфейс логирования.
type Logger interface {
	WithFields(keyValues ...interface{}) Logger
	Error(err error)
	Debug(msg string)
}
