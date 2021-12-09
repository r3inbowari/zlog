package test

import (
	"github.com/fatih/color"
	"github.com/r3inbowari/zlog"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

func TestWithTag(t *testing.T) {
	l := zlog.NewLogger()
	l.WithTag("BSC").WithField("a", uint64(18446744073709551615)).Info("hello, world!")
	l.WithTag("BSC").Info("hello, world!")
	l.WithTag("").Info("hello, world!")
	l.Info("hello, world!")
}

func TestLevel(t *testing.T) {
	l := zlog.NewLogger()
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

func TestRotate(t *testing.T) {
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

func TestMap(t *testing.T) {
	a := map[string]interface{}{
		"111":  "aa",
		"1112": "aa",
		"1232": 123,
	}
	println(zlog.MapToJson(a))
	b := map[string]interface{}{}
	println(zlog.MapToJson(b))
}
