package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	opt := defaultLogOption()
	opt.Format = SimpleColorFormat
	InitLog(opt)
	Info("gogogo")
	Infof("good luck%s %%", "ABC")
	Info("love you", "jjj")
	Info("love %%", "ee")
	Info("love %%%s", "JASON")
	Warning("ok")
	Error("ok")
	Debug("ok")
	Notice("ok")
	Critical("ok")
}
