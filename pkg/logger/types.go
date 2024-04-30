package logger

type LoggerV1 interface {
	Debug(msg string, args ...Filed)
	Info(msg string, args ...Filed)
	Warn(msg string, args ...Filed)
	Error(msg string, args ...Filed)
}

type Filed struct {
	Key   string
	Value any
}
