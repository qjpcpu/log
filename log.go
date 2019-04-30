package log

import (
	"errors"
	"github.com/qjpcpu/filelog"
	"github.com/qjpcpu/log/logging"
	"io"
	syslog "log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
)

type moduleLoggers struct {
	loggers map[string]*logWrapper
	*sync.RWMutex
}

type logWrapper struct {
	*logging.Logger
	option        *LogOption
	leveldBackend logging.LeveledBackend
}

// package global variables
var (
	mloggers   *moduleLoggers
	defaultLgr *logWrapper
)

const (
	// NormFormat without color
	NormFormat = "%{level} %{time:2006-01-02 15:04:05.000} %{shortfile} %{message}"
	// DebugFormat with color
	DebugFormat = "%{level} %{time:2006-01-02 15:04:05.000} grtid:%{goroutineid}/gcnt:%{goroutinecount} %{shortfile} %{message}"
	// SimpleColorFormat simple format with color
	SimpleColorFormat = "\033[1;33m%{level}\033[0m \033[1;36m%{time:2006-01-02 15:04:05.000}\033[0m \033[0;34m%{shortfile}\033[0m \033[0;32m%{message}\033[0m"
	// DebugColorFormat with color
	DebugColorFormat = "\033[1;33m%{level}\033[0m \033[1;36m%{time:2006-01-02 15:04:05.000}\033[0m \033[0;34m%{shortfile}\033[0m \033[0;32mgrtid:%{goroutineid}/gcnt:%{goroutinecount}\033[0m %{message}"
	// CliFormat simple format
	CliFormat = "\033[1;33m%{level}\033[0m \033[1;36m%{time:2006-01-02 15:04:05}\033[0m \033[0;32m%{message}\033[0m"
)

// Level log level
type Level int

const (
	// CRITICAL level
	CRITICAL Level = iota + 1
	// ERROR level
	ERROR
	// WARNING level
	WARNING
	// NOTICE level
	NOTICE
	// INFO level
	INFO
	// DEBUG level
	DEBUG
)

func (lvl Level) loggingLevel() logging.Level {
	return logging.Level(lvl - 1)
}

func parseLogLevel(lstr string) Level {
	lstr = strings.ToLower(lstr)
	switch lstr {
	case "critical":
		return CRITICAL
	case "error":
		return ERROR
	case "warning":
		return WARNING
	case "notice":
		return NOTICE
	case "info":
		return INFO
	case "debug":
		return DEBUG
	default:
		return INFO
	}
}

// LogOption log config options
type LogOption struct {
	LogFile        string
	Level          Level
	Format         string
	RotateType     filelog.RotateType
	CreateShortcut bool
	ErrorLogFile   string
	files          []io.WriteCloser
	module         string
}

// RotateType 轮转类型
type RotateType int

const (
	// RotateDaily 按天轮转
	RotateDaily RotateType = iota
	// RotateHourly 按小时轮转
	RotateHourly
	// RotateWeekly 按周轮转
	RotateWeekly
	// RotateNone 不切割日志
	RotateNone
)

// GetMBuilder module log builder
func GetMBuilder(m string) *LogOption {
	opt := defaultLogOption()
	opt.module = m
	return &opt
}

// GetBuilder log builder
func GetBuilder() *LogOption {
	opt := defaultLogOption()
	return &opt
}

// SetFile set log file
func (lo *LogOption) SetFile(filename string) *LogOption {
	lo.LogFile = filename
	return lo
}

// SetLevel set log level
func (lo *LogOption) SetLevel(level string) *LogOption {
	lo.Level = parseLogLevel(level)
	return lo
}

// SetTypedLevel set log level
func (lo *LogOption) SetTypedLevel(level Level) *LogOption {
	lo.Level = level
	return lo
}

// SetFormat set log format
func (lo *LogOption) SetFormat(format string) *LogOption {
	lo.Format = format
	return lo
}

// SetRotate set rotate type default daily
func (lo *LogOption) SetRotate(rt RotateType) *LogOption {
	lo.RotateType = filelog.RotateType(rt)
	return lo
}

// SetShortcut whether create shorcut when rotate
func (lo *LogOption) SetShortcut(create bool) *LogOption {
	lo.CreateShortcut = create
	return lo
}

// SetErrorLog set error log suffix,default is wf
func (lo *LogOption) SetErrorLog(f string) *LogOption {
	lo.ErrorLogFile = f
	return lo
}

// Submit use this buider options
func (lo *LogOption) Submit() {
	lgr := createLogger(lo)
	if lo.module == "" {
		defaultLgr = lgr
	} else {
		lgr.ExtraCalldepth--
		mloggers.Lock()
		defer mloggers.Unlock()
		mloggers.loggers[lo.module] = lgr
	}
}

// M module log
func M(m string) *logging.Logger {
	mloggers.RLock()
	defer mloggers.RUnlock()
	return mloggers.loggers[m].Logger
}

func defaultLogOption() LogOption {
	return LogOption{
		Level:          DEBUG,
		Format:         DebugColorFormat,
		RotateType:     filelog.RotateNone,
		CreateShortcut: false,
		module:         "",
	}
}

func init() {
	mloggers = &moduleLoggers{
		RWMutex: new(sync.RWMutex),
		loggers: make(map[string]*logWrapper),
	}
	dopt := defaultLogOption()
	defaultLgr = createLogger(&dopt)
}

func createLogger(opt *LogOption) *logWrapper {
	if opt.Format == "" {
		opt.Format = NormFormat
	}
	if opt.Level <= 0 {
		opt.Level = INFO
	}
	lgr := logging.MustGetLogger(opt.module)
	format := logging.MustStringFormatter(opt.Format)

	var leveldBackend logging.LeveledBackend
	if opt.LogFile != "" {
		var backends []logging.LeveledBackend
		// mkdir log dir
		os.MkdirAll(filepath.Dir(opt.LogFile), 0777)
		os.MkdirAll(filepath.Dir(opt.ErrorLogFile), 0777)
		filename := opt.LogFile
		infoLogFp, err := filelog.NewWriter(filename, func(fopt *filelog.Option) {
			fopt.RotateType = opt.RotateType
			fopt.CreateShortcut = opt.CreateShortcut
		})
		if err != nil {
			syslog.Fatalf("open file[%s] failed[%s]", filename, err)
		}
		backendInfo := logging.NewLogBackend(infoLogFp, "", 0)
		backendInfoFormatter := logging.NewBackendFormatter(backendInfo, format)
		backendInfoLeveld := logging.AddModuleLevel(backendInfoFormatter)
		backendInfoLeveld.SetLevel(opt.Level.loggingLevel(), "")
		leveldBackend = backendInfoLeveld
		backends = append(backends, backendInfoLeveld)
		opt.files = append(opt.files, infoLogFp)

		if opt.ErrorLogFile != "" && opt.ErrorLogFile != opt.LogFile {
			errLogFp, err := filelog.NewWriter(opt.ErrorLogFile, func(fopt *filelog.Option) {
				fopt.RotateType = opt.RotateType
				fopt.CreateShortcut = opt.CreateShortcut
			})
			if err != nil {
				syslog.Fatalf("open file[%s.wf] failed[%s]", filename, err)
			}

			backendErr := logging.NewLogBackend(errLogFp, "", 0)
			backendErrFormatter := logging.NewBackendFormatter(backendErr, format)
			backendErrLeveld := logging.AddModuleLevel(backendErrFormatter)
			backendErrLeveld.SetLevel(logging.ERROR, "")
			backends = append(backends, backendErrLeveld)
			opt.files = append(opt.files, errLogFp)
		}
		var bl []logging.Backend
		for _, lb := range backends {
			bl = append(bl, lb)
		}
		ml := logging.MultiLogger(bl...)
		lgr.SetBackend(ml)
	} else {
		backend1 := logging.NewLogBackend(os.Stderr, "", 0)
		backend1Formatter := logging.NewBackendFormatter(backend1, format)
		backend1Leveled := logging.AddModuleLevel(backend1Formatter)
		backend1Leveled.SetLevel(opt.Level.loggingLevel(), "")
		leveldBackend = backend1Leveled

		lgr.SetBackend(backend1Leveled)
	}
	lgr.ExtraCalldepth++
	return &logWrapper{Logger: lgr, option: opt, leveldBackend: leveldBackend}
}

// Infof write leveled log
func Infof(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Infof(format, args...)
}

// Warningf write leveled log
func Warningf(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Warningf(format, args...)
}

// Criticalf write leveled log
func Criticalf(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Criticalf(format, args...)
}

// Fatalf write leveled log
func Fatalf(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Fatalf(format, args...)
}

// Errorf write leveled log
func Errorf(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Errorf(format, args...)
}

// Debugf write leveled log
func Debugf(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Debugf(format, args...)
}

// Noticef write leveled log
func Noticef(format string, args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Noticef(format, args...)
}

// Info write leveled log
func Info(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Infof(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// Warning write leveled log
func Warning(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Warningf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// Critical write leveled log
func Critical(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Criticalf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// Fatal write leveled log
func Fatal(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Fatalf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// Error write leveled log
func Error(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Errorf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// Debug write leveled log
func Debug(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Debugf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// Notice write leveld log
func Notice(args ...interface{}) {
	if defaultLgr == nil {
		return
	}
	defaultLgr.Noticef(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// MustNoErr panic when err occur, should only used in test
func MustNoErr(err error, desc ...string) {
	if err != nil {
		stackInfo := debug.Stack()
		start := 0
		count := 0
		for i, ch := range stackInfo {
			if ch == '\n' {
				if count == 0 {
					start = i
				} else if count == 4 {
					stackInfo = append(stackInfo[0:start+1], stackInfo[i+1:]...)
					break
				}
				count++
			}
		}
		var extra string
		if len(desc) > 0 && desc[0] != "" {
			extra = "[" + desc[0] + "]"
		}
		defaultLgr.Fatalf("%s%v\nMustNoErr fail, %s", extra, err, stackInfo)
	}
}

// GetLogLevel default logger level
func GetLogLevel() string {
	switch defaultLgr.option.Level {
	case CRITICAL:
		return "critical"
	case ERROR:
		return "error"
	case WARNING:
		return "warning"
	case NOTICE:
		return "notice"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	default:
		return ""
	}
}

// SetLogLevel default logger level
func SetLogLevel(lvl string) error {
	tlvl := parseLogLevel(lvl)
	defaultLgr.option.Level = tlvl
	defaultLgr.leveldBackend.SetLevel(tlvl.loggingLevel(), "")
	return nil
}

// SetMLogLevel set module log level
func SetMLogLevel(module, lvl string) error {
	mloggers.RLock()
	defer mloggers.RUnlock()
	wl, ok := mloggers.loggers[module]
	if !ok {
		return errors.New("no such module " + module)
	}
	tlvl := parseLogLevel(lvl)
	wl.option.Level = tlvl
	wl.leveldBackend.SetLevel(tlvl.loggingLevel(), "")
	return nil
}

// GetMLogLevel get module log level
func GetMLogLevel(module string) string {
	mloggers.RLock()
	defer mloggers.RUnlock()
	wl, ok := mloggers.loggers[module]
	if !ok {
		return ""
	}
	switch wl.option.Level {
	case CRITICAL:
		return "critical"
	case ERROR:
		return "error"
	case WARNING:
		return "warning"
	case NOTICE:
		return "notice"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	default:
		return ""
	}
}
