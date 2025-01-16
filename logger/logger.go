package logger

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	SugarLog *zap.SugaredLogger
	LG       *zap.Logger
)

func Init(mode string) (err error) {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte("debug"))
	if err != nil {
		return
	}
	var core zapcore.Core
	if mode == "dev" {
		// 进入开发模式，日志输出到终端,同时也写到文件中
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, l),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		)
	} else {
		core = zapcore.NewCore(encoder, writeSyncer, l)
	}

	LG = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(LG)
	SugarLog = LG.Sugar()
	zap.L().Info("zap.L().info logger success")
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./logs/autoCrateASk.log",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     30,
	}
	return zapcore.AddSync(lumberJackLogger)
}
