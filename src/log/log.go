package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
)

var sugar *zap.SugaredLogger
var pool = buffer.NewPool()

func init() {
	err := zap.RegisterEncoder("custom", constructEncoder)
	if err != nil {
		panic(err)
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		Encoding:         "custom",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	sugar = logger.Sugar()
}

func Infow(msg string, keysAndValues ...interface{}) {
	sugar.Infow(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	sugar.Errorw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	sugar.Fatalw(msg, keysAndValues...)
}
