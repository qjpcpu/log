package log

import (
	"github.com/qjpcpu/log/logging"
	syslog "log"
	"os"
	"path/filepath"
	"strings"
)

// package global variables
var lg *logging.Logger
var setLogLevel func(Level)
var log_option = defaultLogOption()

const (
	NormFormat = "%{level}: [%{time:2006-01-02 15:04:05.000}][%{goroutineid}/%{goroutinecount}][%{shortfile}][%{message}]"
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

type LogOption struct {
	LogFile    string
	Level      Level
	Format     string
	RotateType logging.RotateType
}

func defaultLogOption() LogOption {
	return LogOption{
		Level:      DEBUG,
		Format:     NormFormat,
		RotateType: logging.RotateDaily,
	}
}

func init() {
	InitLog(defaultLogOption())
}

func InitLog(opt LogOption) {
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
		filename := opt.LogFile
		info_log_fp, err := logging.NewFileLogWriter(filename, opt.RotateType)
		if err != nil {
			syslog.Fatalf("open file[%s] failed[%s]", filename, err)
		}

		err_log_fp, err := logging.NewFileLogWriter(filename+".wf", opt.RotateType)
		if err != nil {
			syslog.Fatalf("open file[%s.wf] failed[%s]", filename, err)
		}

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
	lg.ExtraCalldepth += 1
	log_option = opt
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

func SetLogLevel(lvl Level) {
	if setLogLevel != nil {
		setLogLevel(lvl)
		log_option.Level = lvl
	}
}

func GetLogLevel() Level {
	return log_option.Level
}
