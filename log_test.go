package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	InitLog(CliLogOption())
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
