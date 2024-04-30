package logger

import "go.uber.org/zap"

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(l *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: l,
	}
}

func (z *ZapLogger) Debug(msg string, args ...Filed) {
	z.logger.Debug(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Info(msg string, args ...Filed) {
	z.logger.Info(msg, z.toArgs(args)...)

}

func (z *ZapLogger) Warn(msg string, args ...Filed) {
	z.logger.Warn(msg, z.toArgs(args)...)
}

func (z *ZapLogger) Error(msg string, args ...Filed) {
	z.logger.Error(msg, z.toArgs(args)...)
}

func (z *ZapLogger) toArgs(args []Filed) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}
