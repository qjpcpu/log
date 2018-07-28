package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	Info("gogogo")
	Infof("good luck%s %%", "ABC")
	Info("love you", "jjj")
	Info("love %%", "ee")
	Info("love %%%s", "JASON")
}
