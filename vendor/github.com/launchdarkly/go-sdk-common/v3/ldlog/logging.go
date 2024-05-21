package ldlog

import (
	"log"
	"os"
	"strings"
)

// BaseLogger is a generic logger interface with no level mechanism. Since its methods are a
// subset of Go's [log.Logger], you may use [log.New] to create a BaseLogger.
//
// This is identical to the Logger interface in the main SDK package. It is redefined here so
// the ldlog package does not have to refer back to the main package.
type BaseLogger interface {
	// Logs a message on a single line. This is equivalent to log.Logger.Println.
	Println(values ...interface{})
	// Logs a message on a single line, applying a format string. This is equivalent to log.Logger.Printf.
	Printf(format string, values ...interface{})
}

// LogLevel describes one of the possible thresholds of log message, from LogDebug to LogError.
type LogLevel int

// Name returns a descriptive name for this log level.
func (level LogLevel) Name() string {
	switch level {
	case Debug:
		return "Debug"
	case Info:
		return "Info"
	case Warn:
		return "Warn"
	case Error:
		return "Error"
	case None:
		return "None"
	}
	return "?"
}

// String is the default string representation of LogLevel, which is the same as [LogLevel.Name].
func (level LogLevel) String() string {
	return level.Name()
}

const (
	_ = iota
	// Debug is the least significant logging level, containing verbose output you will normally
	// not need to see. This level is disabled by default.
	Debug LogLevel = iota
	// Info is the logging level for informational messages about normal operations. This level
	// is enabled by default.
	Info LogLevel = iota
	// Warn is the logging level for more significant messages about an uncommon condition that
	// is not necessarily an error. This level is enabled by default.
	Warn LogLevel = iota
	// Error is the logging level for error conditions that should not happen during normal
	// operation of the SDK. This level is enabled by default.
	Error LogLevel = iota
	// None means no messages at all should be logged.
	None LogLevel = iota
)

// Loggers is a configurable logging component with a level filter.
//
// By default, Loggers sends output to standard error and enables all levels except [Debug].
// You may call any of its Set methods to change this configuration.
type Loggers struct {
	debugLog   levelLogger
	infoLog    levelLogger
	warnLog    levelLogger
	errorLog   levelLogger
	baseLogger BaseLogger
	minLevel   LogLevel
	prefix     string
	inited     bool
}

type levelLogger struct {
	baseLogger     BaseLogger
	enabled        bool
	prefix         string
	overrideLogger bool
}

var nullLog = levelLogger{enabled: false} //nolint:gochecknoglobals

// NewDisabledLoggers returns a Loggers instance that will never generate output.
func NewDisabledLoggers() Loggers {
	ret := Loggers{}
	ret.SetBaseLogger(nullLog)
	ret.SetMinLevel(None)
	return ret
}

// Debug logs a message at Debug level, if that level is enabled. It calls the BaseLogger's Println.
func (l Loggers) Debug(values ...interface{}) {
	l.ForLevel(Debug).Println(values...)
}

// Debugf logs a message at Debug level with a format string, if that level is enabled. It calls the
// BaseLogger's Printf.
func (l Loggers) Debugf(format string, values ...interface{}) {
	l.ForLevel(Debug).Printf(format, values...)
}

// Info logs a message at Info level, if that level is enabled. It calls the BaseLogger's Println.
func (l Loggers) Info(values ...interface{}) {
	l.ForLevel(Info).Println(values...)
}

// Infof logs a message at Info level with a format string, if that level is enabled. It calls the
// BaseLogger's Printf.
func (l Loggers) Infof(format string, values ...interface{}) {
	l.ForLevel(Info).Printf(format, values...)
}

// Warn logs a message at Warn level, if that level is enabled. It calls the BaseLogger's Println.
func (l Loggers) Warn(values ...interface{}) {
	l.ForLevel(Warn).Println(values...)
}

// Warnf logs a message at Warn level with a format string, if that level is enabled. It calls the
// BaseLogger's Printf.
func (l Loggers) Warnf(format string, values ...interface{}) {
	l.ForLevel(Warn).Printf(format, values...)
}

// Error logs a message at Error level, if that level is enabled. It calls the BaseLogger's Println.
func (l Loggers) Error(values ...interface{}) {
	l.ForLevel(Error).Println(values...)
}

// Errorf logs a message at Error level with a format string, if that level is enabled. It calls the
// BaseLogger's Printf.
func (l Loggers) Errorf(format string, values ...interface{}) {
	l.ForLevel(Error).Printf(format, values...)
}

// NewDefaultLoggers returns a new Loggers instance with default properties.
//
// This is different from an empty Loggers{} instance in that the latter will not produce any output
// until you call either [Loggers.SetBaseLogger], [Loggers.SetMinLevel], or [Loggers.Init] on it.
// Calling NewDefaultLoggers() ensures that the default minimum level of [Info] is enabled.
func NewDefaultLoggers() Loggers {
	ret := Loggers{}
	ret.SetMinLevel(Info)
	return ret
}

// Init ensures that the Loggers instance is ready to use.
//
// This is necessary only if you have a Loggers instance that was not produced by [NewDefaultLoggers]
// and that may not have had [Loggers.SetBaseLogger] or [Loggers.SetMinLevel] called on it. It ensures
// that the default properties have been set. If you have already set any properties, Init does nothing.
func (l *Loggers) Init() {
	l.ensureInited()
}

// SetBaseLogger specifies the default destination for output at all log levels. This does not apply
// to any levels whose BaseLogger has been overridden with [Loggers.SetBaseLoggerForLevel]. All messages
// written to this logger will be prefixed with "NAME: " where NAME is DEBUG, INFO, etc.
//
// If baseLogger is nil, nothing is changed.
func (l *Loggers) SetBaseLogger(baseLogger BaseLogger) {
	l.ensureInited()
	if baseLogger == nil {
		return
	}
	l.baseLogger = baseLogger
	for _, levelLogger := range l.allLevels() {
		if !levelLogger.overrideLogger {
			levelLogger.baseLogger = baseLogger
		}
	}
}

// SetBaseLoggerForLevel specifies the default destination for output at the given log level. All
// messages written to this logger will be prefixed with "NAME: " where NAME is DEBUG, INFO, etc.
//
// If baseLogger is nil, this level will use the default from [Loggers.SetBaseLogger].
func (l *Loggers) SetBaseLoggerForLevel(level LogLevel, baseLogger BaseLogger) {
	l.ensureInited()
	levelLogger := l.levelLogger(level)
	if levelLogger != nil {
		if baseLogger == nil {
			levelLogger.baseLogger = l.baseLogger
			levelLogger.overrideLogger = false
		} else {
			levelLogger.baseLogger = baseLogger
			levelLogger.overrideLogger = true
		}
	}
}

// ForLevel returns a BaseLogger that writes messages at the specified level. Use this if you have
// code that already uses the Printf/Println methods. All of the existing level configuration still
// applies, so, for instance, loggers.ForLevel(Debug).Println("x") is exactly the same as
// loggers.Debug("x").
//
// If the level is not a valid log level, the return value is non-nil but will produce no output.
func (l Loggers) ForLevel(level LogLevel) BaseLogger {
	if level >= l.minLevel {
		lll := l.levelLogger(level)
		if lll != nil {
			return *lll
		}
	}
	return nullLog
}

// SetMinLevel specifies the minimum level for log output, where [Debug] is the lowest and [Error]
// is the highest. Log messages at a level lower than this will be suppressed. The default is
// [Info].
func (l *Loggers) SetMinLevel(minLevel LogLevel) {
	l.ensureInited()
	l.minLevel = minLevel
	l.configureLevels()
}

// GetMinLevel returns the minimum level that has been specified for log output. The default is [Info].
func (l Loggers) GetMinLevel() LogLevel {
	if l.minLevel == 0 {
		return Info // this instance hasn't been initialized, use the default
	}
	return l.minLevel
}

// IsDebugEnabled returns true if the minimum log level is [Debug], or false if it is higher.
//
// This allows for greater efficiency in code that can produce verbose debug output. When the Debug
// level is disabled, calling [Loggers.Debug] or [Loggers.Debugf] does not produce any output but they
// can still cause unwanted overhead due to having to convert their parameters to interface{} values. To
// avoid that overhead, you can choose to not bother calling Debug or Debugf at all if IsDebugEnabled
// returns false.
func (l Loggers) IsDebugEnabled() bool {
	return l.GetMinLevel() == Debug
}

// SetPrefix specifies a string to be added before every log message, after the LEVEL: prefix.
// Do not include a trailing space.
func (l *Loggers) SetPrefix(prefix string) {
	l.ensureInited()
	l.prefix = prefix
	l.configureLevels()
}

func (l *Loggers) ensureInited() {
	if l.inited {
		return
	}
	l.minLevel = Info
	l.baseLogger = log.New(os.Stderr, "[LaunchDarkly] ", log.LstdFlags)
	for _, levelLogger := range l.allLevels() {
		levelLogger.baseLogger = l.baseLogger
	}
	l.configureLevels()
	l.inited = true
}

func (l *Loggers) configureLevels() {
	for level, levelLogger := range l.allLevels() {
		levelLogger.enabled = level >= l.minLevel
		levelLogger.prefix = strings.ToUpper(level.Name()) + ":"
		if l.prefix != "" {
			levelLogger.prefix = levelLogger.prefix + " " + l.prefix
		}
	}
}

func (l *Loggers) allLevels() map[LogLevel]*levelLogger {
	return map[LogLevel]*levelLogger{
		Debug: &l.debugLog,
		Info:  &l.infoLog,
		Warn:  &l.warnLog,
		Error: &l.errorLog,
	}
}

func (l *Loggers) levelLogger(level LogLevel) *levelLogger {
	switch level {
	case Debug:
		return &l.debugLog
	case Info:
		return &l.infoLog
	case Warn:
		return &l.warnLog
	case Error:
		return &l.errorLog
	}
	return nil
}

func (ll levelLogger) Println(values ...interface{}) {
	if ll.enabled && ll.baseLogger != nil {
		if len(values) == 1 {
			ll.baseLogger.Println(ll.prefix, values[0])
		} else {
			vs := make([]interface{}, len(values)+1)
			vs[0] = ll.prefix
			for i := range values {
				vs[i+1] = values[i]
			}
			ll.baseLogger.Println(vs...)
		}
	}
}

func (ll levelLogger) Printf(format string, args ...interface{}) {
	if ll.enabled && ll.baseLogger != nil {
		ll.baseLogger.Printf(ll.prefix+" "+format, args...)
	}
}
