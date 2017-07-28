package log

import (
	"github.com/qjpcpu/log/logging"
	syslog "log"
	"os"
)

const (
	NormFormat = "%{level}: [%{time:2006-01-02 15:04:05.000}][%{pid}][%{goroutineid}/%{goroutinecount}][%{shortfile}][%{message}]"
)

type Level int

const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

type LogOption struct {
	LogFile    string
	LogDir     string
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

var lg *logging.Logger

func init() {
	InitLog(defaultLogOption())
}

func InitLog(opt LogOption) {
	if opt.Format == "" {
		opt.Format = NormFormat
	}
	format := logging.MustStringFormatter(opt.Format)
	// mkdir log dir
	if opt.LogDir != "" && opt.LogFile != "" {
		os.MkdirAll(opt.LogDir, 0777)
		filename := opt.LogDir + "/" + opt.LogFile
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
		backend_info_leveld.SetLevel(logging.Level(opt.Level), "")

		backend_err_leveld := logging.AddModuleLevel(backend_err_formatter)
		backend_err_leveld.SetLevel(logging.ERROR, "")
		logging.SetBackend(backend_info_leveld, backend_err_leveld)

	} else {
		backend1 := logging.NewLogBackend(os.Stderr, "", 0)
		backend2 := logging.NewLogBackend(os.Stderr, "", 0)
		backend2Formatter := logging.NewBackendFormatter(backend2, format)
		backend1Leveled := logging.AddModuleLevel(backend1)
		backend1Leveled.SetLevel(logging.Level(opt.Level), "")
		logging.SetBackend(backend1Leveled, backend2Formatter)
	}
	lg = logging.MustGetLogger("")
	lg.ExtraCalldepth += 1
}

func Info(format string, args ...interface{}) {
	if lg == nil {
		return
	}
	lg.Infof(format, args...)
}
