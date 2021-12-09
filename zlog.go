package zlog

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	rotate "github.com/r3inbowari/zlog/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var LevelArray = []string{
	// PanicLevel level, the highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	"P",
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	"F",
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	"E",
	// WarnLevel level. Non-critical entries that deserve eyes.
	"W",
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	"I",
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	"D",
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	"T",
}

type Color map[string]color.Attribute

var defaultLevelColor = Color{
	"P": color.FgRed,
	"F": color.FgRed,
	"E": color.FgRed,
	"W": color.FgYellow,
	"I": color.FgGreen,
	"D": color.FgMagenta,
	"T": color.FgHiWhite,
}

type ZLog struct {
	logrus.Logger
	BuildMode    string
	Rotate       *rotate.RotateLogs
	RotateEnable bool
	fm           logrus.MutexWrap
	levelColor   Color
}

func NewLogger() *ZLog {
	var z ZLog
	z.levelColor = defaultLevelColor
	z.SetExitFunc(os.Exit)
	z.SetBuildMode("rel")
	z.SetLevel(logrus.DebugLevel)
	z.SetFormatter(&z)
	z.SetOutput(&z)
	z.SetReportCaller(true)
	return &z
}

func (f *ZLog) SetExitFunc(fn func(i int)) *ZLog {
	f.ExitFunc = fn
	return f
}

func (f *ZLog) SetBuildMode(buildMode string) *ZLog {
	f.BuildMode = strings.ToLower(buildMode)
	return f
}

func (f *ZLog) SetLevelColor(level logrus.Level, attribute color.Attribute) *ZLog {
	f.fm.Lock()
	defer f.fm.Unlock()
	f.levelColor[LevelArray[level%7]] = attribute
	return f
}

func (f *ZLog) SetRotate(rotateEnable bool) *ZLog {
	f.fm.Lock()
	defer f.fm.Unlock()
	if rotateEnable {
		p, err := os.Executable()
		if err != nil {
			return f
		}
		p = filepath.Dir(p) + "\\log\\"
		if rotateEnable {
			writer, _ := rotate.New(
				p+"%Y%m%d%H%M.log",
				rotate.WithLinkName(p),
				rotate.WithMaxAge(time.Duration(180)*time.Second),
				rotate.WithRotationTime(time.Duration(60)*time.Second),
			)
			f.Rotate = writer
		}
	} else {
		f.Rotate = nil
	}
	f.RotateEnable = rotateEnable
	return f
}

func (f *ZLog) Blue(msg string) {
	if f.BuildMode == "rel" {
		color.Blue(msg)
	} else {
		fmt.Printf("\x1b[%dm"+msg+" \x1b[0m\n", 34)
	}
}

// Write implement the Output Writer interface
func (f *ZLog) Write(p []byte) (n int, err error) {
	if f.BuildMode == "rel" {
		n, err = color.New(f.levelColor[string(p[1])]).Println(string(p))
	} else {
		n, err = fmt.Printf("\x1b[%dm"+string(p)+" \x1b[0m\n", f.levelColor[string(p[1])])
	}
	if !f.RotateEnable {
		return
	}
	return f.Rotate.Write(p)
}

// Format implement the Formatter interface
func (f *ZLog) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	remained := len(entry.Data)
	if remained > 0 {
		entry.Message += " ["
	}
	for k, v := range entry.Data {
		entry.Message += k + ":" + fieldParse(v)
		remained--
		if remained != 0 {
			entry.Message += ", "
		} else {
			entry.Message += "]"
		}
	}
	filename := path.Base(entry.Caller.File)
	b.WriteString(fmt.Sprintf("[%s] %s [%s:%d] %s", LevelArray[entry.Level], entry.Time.Format("2006-01-02 15:04:05"), filename, entry.Caller.Line, entry.Message))
	return b.Bytes(), nil
}

func fieldParse(obj interface{}) string {
	var ret string
	switch v := obj.(type) {
	case string:
		ret = v
	case float64:
		ret = strconv.FormatFloat(v, 'E', -1, 64)
	case int:
		ret = strconv.Itoa(v)
	case int64:
		ret = strconv.FormatInt(v, 0x0a)
	case nil:
		ret = "nil"
	default:
		ret = "unsupported"
	}
	return ret
}

var Log *ZLog

func InitGlobalLogger() *ZLog {
	if Log == nil {
		Log = NewLogger()
	}
	return Log
}
