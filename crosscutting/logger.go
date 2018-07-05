package crosscutting

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	l       *logger
	onceLog sync.Once
)

type logger struct {
	*zap.Logger
}

func Logger() *logger {
	onceLog.Do(func() {

		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		cfg.OutputPaths = Config().GetStringSlice("app.log.out")
		fmt.Println("Application Logger initialisation")
		log, err := cfg.Build()
		if err != nil {
			panic(err)
		}
		defer log.Sync()
		l = &logger{log}
	})
	return l
}

func (l *logger) Error(err error) zap.Field {
	return zap.Error(err)
}
