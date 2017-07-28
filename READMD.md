```
package main

import (
	"github.com/qjpcpu/log"
	"time"
)

func main() {
	log.InitLog(log.LogOption{
		LogDir:  "./log",
		Level:   log.DEBUG,
		LogFile: "access.log",
	})
	log.Info("GO %s", "log")
	for i := 0; i < 100; i++ {
		log.Info("time:%v", time.Now())
		time.Sleep(30 * time.Second)
	}
}
```
