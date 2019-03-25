package log

import (
	"github.com/qjpcpu/filelog"
	"github.com/qjpcpu/log/logging"
	"io"
	syslog "log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// package global variables
var lg *logging.Logger
var setLogLevel func(Level)
var logOption = defaultLogOption()

const (
	NormFormat        = "%{level} %{time:2006-01-02 15:04:05.000} %{shortfile} %{message}"
	DebugFormat       = "%{level} %{time:2006-01-02 15:04:05.000} grtid:%{goroutineid}/gcnt:%{goroutinecount} %{shortfile} %{message}"
	SimpleColorFormat = "\033[1;33m%{level}\033[0m \033[1;36m%{time:2006-01-02 15:04:05.000}\033[0m \033[0;34m%{shortfile}\033[0m \033[0;32m%{message}\033[0m"
	DebugColorFormat  = "\033[1;33m%{level}\033[0m \033[1;36m%{time:2006-01-02 15:04:05.000}\033[0m \033[0;34m%{shortfile}\033[0m \033[0;32mgrtid:%{goroutineid}/gcnt:%{goroutinecount}\033[0m %{message}"
	CliFormat         = "\033[1;33m%{level}\033[0m \033[1;36m%{time:2006-01-02 15:04:05}\033[0m \033[0;32m%{message}\033[0m"
)

type Level int

const (
	CRITICAL Level = iota + 1
	ERROR
	WARNING
	NOTICE
	INFO
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
	if lo.ErrorLogFile == "" {
		lo.ErrorLogFile = lo.LogFile + ".error"
	}
	initLog(*lo)
}

func defaultLogOption() LogOption {
	return LogOption{
		Level:          DEBUG,
		Format:         DebugColorFormat,
		RotateType:     filelog.RotateNone,
		CreateShortcut: false,
	}
}

func init() {
	initLog(defaultLogOption())
}

func initLog(opt LogOption) {
	if len(logOption.files) > 0 {
		for _, f := range logOption.files {
			if f != nil {
				f.Close()
			}
		}
		logOption.files = nil
	}
	if opt.Format == "" {
		opt.Format = NormFormat
	}
	if opt.Level <= 0 {
		opt.Level = INFO
	}
	format := logging.MustStringFormatter(opt.Format)
	if opt.LogFile != "" {
		// mkdir log dir
		os.MkdirAll(filepath.Dir(opt.LogFile), 0777)
		os.MkdirAll(filepath.Dir(opt.ErrorLogFile), 0777)
		filename := opt.LogFile
		info_log_fp, err := filelog.NewWriter(filename, func(fopt *filelog.Option) {
			fopt.RotateType = opt.RotateType
			fopt.CreateShortcut = opt.CreateShortcut
		})
		if err != nil {
			syslog.Fatalf("open file[%s] failed[%s]", filename, err)
		}

		err_log_fp, err := filelog.NewWriter(opt.ErrorLogFile, func(fopt *filelog.Option) {
			fopt.RotateType = opt.RotateType
			fopt.CreateShortcut = opt.CreateShortcut
		})
		if err != nil {
			syslog.Fatalf("open file[%s.wf] failed[%s]", filename, err)
		}
		opt.files = []io.WriteCloser{info_log_fp, err_log_fp}

		backend_info := logging.NewLogBackend(info_log_fp, "", 0)
		backend_err := logging.NewLogBackend(err_log_fp, "", 0)
		backend_info_formatter := logging.NewBackendFormatter(backend_info, format)
		backend_err_formatter := logging.NewBackendFormatter(backend_err, format)

		backend_info_leveld := logging.AddModuleLevel(backend_info_formatter)
		backend_info_leveld.SetLevel(opt.Level.loggingLevel(), "")

		backend_err_leveld := logging.AddModuleLevel(backend_err_formatter)
		backend_err_leveld.SetLevel(logging.ERROR, "")
		logging.SetBackend(backend_info_leveld, backend_err_leveld)

		// set log level handler
		setLogLevel = func(lvl Level) {
			backend_info_leveld.SetLevel(lvl.loggingLevel(), "")
		}
	} else {
		backend1 := logging.NewLogBackend(os.Stderr, "", 0)
		backend1Formatter := logging.NewBackendFormatter(backend1, format)
		backend1Leveled := logging.AddModuleLevel(backend1Formatter)
		backend1Leveled.SetLevel(opt.Level.loggingLevel(), "")
		logging.SetBackend(backend1Leveled)
		// set log level handler
		setLogLevel = func(lvl Level) {
			backend1Leveled.SetLevel(lvl.loggingLevel(), "")
		}
	}
	lg = logging.MustGetLogger("")
	lg.ExtraCalldepth++
	logOption = opt
}

func Infof(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Infof(format, args...)
}

func Warningf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Warningf(format, args...)
}

func Criticalf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Criticalf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Fatalf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Errorf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Debugf(format, args...)
}

func Noticef(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Noticef(format, args...)
}

func Info(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Infof(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

func Warning(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Warningf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

func Critical(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Criticalf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

func Fatal(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Fatalf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

func Error(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Errorf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

func Debug(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Debugf(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

func Notice(args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Noticef(strings.TrimSpace(strings.Repeat("%+v ", len(args))), args...)
}

// MustNoErr panic when err occur, should only used in test
func MustNoErr(err error, desc ...string) {
	if err != nil {
		stack_info := debug.Stack()
		start := 0
		count := 0
		for i, ch := range stack_info {
			if ch == '\n' {
				if count == 0 {
					start = i
				} else if count == 4 {
					stack_info = append(stack_info[0:start+1], stack_info[i+1:]...)
					break
				}
				count++
			}
		}
		var extra string
		if len(desc) > 0 && desc[0] != "" {
			extra = "[" + desc[0] + "]"
		}
		lg.Fatalf("%s%v\nMustNoErr fail, %s", extra, err, stack_info)
	}
}

// SetLogLevel dynamic set log level
func SetLogLevel(lvl Level) {
	if setLogLevel != nil {
		setLogLevel(lvl)
		logOption.Level = lvl
	}
}

// GetLogLevel get current log level
func GetLogLevel() Level {
	return logOption.Level
}

// Close close log file
func Close() {
	for _, wc := range logOption.files {
		if wc != nil {
			wc.Close()
		}
	}
}
