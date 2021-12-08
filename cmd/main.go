package main

import (
	"github.com/r3inbowari/zlog"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	l := zlog.NewLogger()
	l.SetBuildMode("rel")
	l.SetRotate(true)
	l.SetExitFunc(os.Exit)
	l.SetLevel(logrus.TraceLevel)
	l.Info("hello, world!")
}
