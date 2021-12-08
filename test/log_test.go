package test

import (
	"github.com/fatih/color"
	"github.com/r3inbowari/zlog"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	l := zlog.NewLogger()
	l.SetBuildMode("rel")
	l.SetRotate(false)
	l.SetExitFunc(os.Exit)
	l.SetLevel(logrus.TraceLevel)

	l.Error("hello, world!")
	l.Warn("hello, world!")
	l.Info("hello, world!")
	l.Debug("hello, world!")
	l.Trace("hello, world!")
	l.SetLevelColor(logrus.DebugLevel, color.FgRed)
	l.Debug("hello, world!")
}

func TestB(t *testing.T) {
	l := zlog.NewLogger()
	l.SetBuildMode("rel")
	l.SetRotate(false)
	l.SetExitFunc(os.Exit)
	l.SetLevel(logrus.TraceLevel)
	l.WithField("hello", "world").Info("Hello")
}

func TestC(t *testing.T) {
	l := zlog.NewLogger()
	l.Info("aaaa")
	l.SetRotate(true)

	go func() {
		for i := 0; i < 10000; i++ {
			l.SetRotate(true)
		}
		println("ok")
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			l.SetRotate(false)
		}
		println("ok")
	}()

	time.Sleep(time.Second * 1)
	println("ok")
	l.SetRotate(true)
	l.Info("ok")
}

func TestGlobalLog(t *testing.T) {
	zlog.InitGlobalLogger()
	zlog.Log.Error("hello")
}
