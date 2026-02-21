package logger

import (
	"context"

	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level, format string) *Logger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         format,
		EncoderConfig:    getEncoderConfig(format),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{logger.Sugar()}
}

func getEncoderConfig(format string) zapcore.EncoderConfig {
	if format == "json" {
		return zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}

	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func (l *Logger) Sync() {
	l.SugaredLogger.Sync()
}

// UnaryServerInterceptor returns a gRPC interceptor for logging
func (l *Logger) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract request metadata
		md, _ := metadata.FromIncomingContext(ctx)
		requestID := getMetadataValue(md, "x-request-id")

		// Extract peer info
		peerInfo, _ := peer.FromContext(ctx)
		ipAddress := ""
		if peerInfo != nil {
			ipAddress = peerInfo.Addr.String()
		}

		l.Infow("gRPC request started",
			"method", info.FullMethod,
			"request_id", requestID,
			"ip", ipAddress,
		)

		// Call handler
		resp, err := handler(ctx, req)

		// Log result
		duration := time.Since(start)
		if err != nil {
			l.Errorw("gRPC request failed",
				"method", info.FullMethod,
				"request_id", requestID,
				"duration", duration,
				"error", err,
			)
		} else {
			l.Infow("gRPC request completed",
				"method", info.FullMethod,
				"request_id", requestID,
				"duration", duration,
			)
		}

		return resp, err
	}
}

func getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
