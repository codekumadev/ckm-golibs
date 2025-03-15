package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger := log.New("info", "console", "app.log") // log ลงไฟล์แบบ console
// logger.Info("Logging to file!", log.String("module", "main"))

// logger2 := log.New("debug", "json", "debug.log") // log ลงไฟล์แบบ json
// logger2.Debug("Debug message", log.String("module", "debug"))

type Level = zapcore.Level

const (
	InfoLevel   Level = zap.InfoLevel   // 0, default level
	WarnLevel   Level = zap.WarnLevel   // 1
	ErrorLevel  Level = zap.ErrorLevel  // 2
	DPanicLevel Level = zap.DPanicLevel // 3, used in development log
	// PanicLevel logs a message, then panics
	PanicLevel Level = zap.PanicLevel // 4
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = zap.FatalLevel // 5
	DebugLevel Level = zap.DebugLevel // -1
)

type Field = zap.Field

func (l *Logger) Debug(msg string, fields ...Field) {
	l.l.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.l.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.l.Error(msg, fields...)
}

func (l *Logger) DPanic(msg string, fields ...Field) {
	l.l.DPanic(msg, fields...)
}
func (l *Logger) Panic(msg string, fields ...Field) {
	l.l.Panic(msg, fields...)
}
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, fields...)
}

func (l *Logger) WithFields(fields ...Field) *Logger {
	newLog := l.l.With(fields...)

	return &Logger{
		l:     newLog,
		level: l.level,
	}
}

func (l *Logger) WithOptions(opts ...Option) *Logger {
	newLog := l.l.WithOptions(opts...)
	return &Logger{
		l:     newLog,
		level: l.level,
	}
}

var (
	Skip        = zap.Skip
	Binary      = zap.Binary
	Bool        = zap.Bool
	Boolp       = zap.Boolp
	ByteString  = zap.ByteString
	Complex128  = zap.Complex128
	Complex128p = zap.Complex128p
	Complex64   = zap.Complex64
	Complex64p  = zap.Complex64p
	Float64     = zap.Float64
	Float64p    = zap.Float64p
	Float32     = zap.Float32
	Float32p    = zap.Float32p
	Int         = zap.Int
	Intp        = zap.Intp
	Int64       = zap.Int64
	Int64p      = zap.Int64p
	Int32       = zap.Int32
	Int32p      = zap.Int32p
	Int16       = zap.Int16
	Int16p      = zap.Int16p
	Int8        = zap.Int8
	Int8p       = zap.Int8p
	String      = zap.String
	Stringp     = zap.Stringp
	Uint        = zap.Uint
	Uintp       = zap.Uintp
	Uint64      = zap.Uint64
	Uint64p     = zap.Uint64p
	Uint32      = zap.Uint32
	Uint32p     = zap.Uint32p
	Uint16      = zap.Uint16
	Uint16p     = zap.Uint16p
	Uint8       = zap.Uint8
	Uint8p      = zap.Uint8p
	Uintptr     = zap.Uintptr
	Uintptrp    = zap.Uintptrp
	Reflect     = zap.Reflect
	Namespace   = zap.Namespace
	Stringer    = zap.Stringer
	Time        = zap.Time
	Timep       = zap.Timep
	Stack       = zap.Stack
	StackSkip   = zap.StackSkip
	Duration    = zap.Duration
	Durationp   = zap.Durationp
	Any         = zap.Any
)

type Logger struct {
	l     *zap.Logger // zap ensure that zap.Logger is safe for concurrent use
	level Level
}

type Option = zap.Option

var (
	WithCaller    = zap.WithCaller
	AddStacktrace = zap.AddStacktrace
	AddCallerSkip = zap.AddCallerSkip
	AddCaller     = zap.AddCaller
)

func NewFile(logLevel, logFormat string, logPrefix string) *Logger {
	writer := getLogWriter(logPrefix) // ใช้ไฟล์ที่มีชื่อไม่ซ้ำ
	encoder := zapEncoder(logFormat)
	level := zapLogLevel(logLevel)

	return new(writer, encoder, level, WithCaller(true), AddCallerSkip(2))
}

func New(logLevel, logFormat string) *Logger {
	return new(os.Stderr, zapEncoder(logFormat), zapLogLevel(logLevel), WithCaller(true), AddCallerSkip(2))
}

// ฟังก์ชันในการสร้างชื่อไฟล์ที่ไม่ซ้ำกัน
func generateUniqueLogFileName(prefix string) string {
	// สร้าง timestamp สำหรับชื่อไฟล์
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	// สร้าง UUID เพื่อให้ชื่อไฟล์ไม่ซ้ำ
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%s_%s_%s.log", prefix, timestamp, uniqueID)
}

// ฟังก์ชันนี้จะใช้ชื่อไฟล์ที่ไม่ซ้ำกัน
func getLogWriter(logPrefix string) zapcore.WriteSyncer {
	// สร้างชื่อไฟล์ที่ไม่ซ้ำกัน
	logFileName := generateUniqueLogFileName(logPrefix)

	// เปิดไฟล์สำหรับเขียน log
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file: %v", err))
	}

	return zapcore.AddSync(file)
}

func zapLogLevel(level string) Level {
	switch strings.ToLower(level) {
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpanic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	case "debug":
		return DebugLevel
	}
	return InfoLevel
}

func zapEncoder(logFormat string) zapcore.Encoder {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}

	if strings.ToLower(logFormat) == "console" {
		cfg.Encoding = "console"
		return zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	}

	return zapcore.NewJSONEncoder(cfg.EncoderConfig)
}

func new(writer io.Writer, encoder zapcore.Encoder, level Level, opts ...Option) *Logger {
	if writer == nil {
		panic("the writer is nil")
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		level,
	)
	logger := &Logger{
		l:     zap.New(core, opts...),
		level: level,
	}
	return logger
}

func (l *Logger) Sync() error {
	return l.l.Sync()
}
