package zlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	rotate "github.com/r3inbowari/zlog/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
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

// ZLog Todo implement a option?
type ZLog struct {
	logrus.Logger
	BuildMode       string
	Rotate          *rotate.RotateLogs
	RotateEnable    bool
	fm              logrus.MutexWrap
	levelColor      Color
	ScreenEnable    bool
	WebsocketConn   *websocket.Conn
	WebsocketEnable bool
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

	r := gin.Default()
	r.GET("/log", logBuild(&z))
	go r.Run(":6564")

	return &z
}

func logBuild(l *ZLog) gin.HandlerFunc {
	var upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return func(context *gin.Context) {
		conn, err := upGrader.Upgrade(context.Writer, context.Request, nil)
		if err != nil {
			return
		}
		l.SetWebsocket(conn)
	}
}

func (z *ZLog) SetExitFunc(fn func(i int)) *ZLog {
	z.ExitFunc = fn
	return z
}

func (z *ZLog) SetBuildMode(buildMode string) *ZLog {
	z.BuildMode = strings.ToLower(buildMode)
	return z
}

func (z *ZLog) SetLevelColor(level logrus.Level, attribute color.Attribute) *ZLog {
	z.fm.Lock()
	defer z.fm.Unlock()
	z.levelColor[LevelArray[level%7]] = attribute
	return z
}

func (z *ZLog) SetWebsocket(conn *websocket.Conn) *ZLog {
	z.fm.Lock()
	defer z.fm.Unlock()
	z.WebsocketEnable = false
	z.WebsocketConn = conn
	z.WebsocketEnable = true
	return z
}

func (z *ZLog) SetRotate(rotateEnable bool) *ZLog {
	z.fm.Lock()
	defer z.fm.Unlock()
	if rotateEnable {
		//p, err := os.Executable()
		//if err != nil {
		//	return z
		//}
		//p = filepath.Dir(p) + "\\log\\"
		p := ".\\log\\"

		if rotateEnable {
			writer, _ := rotate.New(
				p+"%Y%m%d%H%M.log",
				rotate.WithLinkName(p),
				rotate.WithMaxAge(time.Duration(180)*time.Second),
				rotate.WithRotationTime(time.Duration(60)*time.Second),
			)
			z.Rotate = writer
		}
	} else {
		z.Rotate = nil
	}
	z.RotateEnable = rotateEnable
	return z
}

func (z *ZLog) SetScreen(screenEnable bool) *ZLog {
	z.ScreenEnable = screenEnable
	return z
}

func (z *ZLog) Blue(msg string) {
	if z.BuildMode == "rel" {
		color.Blue(msg)
	} else {
		fmt.Printf("\x1b[%dm"+msg+" \x1b[0m\n", 34)
	}
}

func (z *ZLog) WithTag(tag string) *logrus.Entry {
	return z.WithField("aGVsbG8=", tag)
}

// Write implement the Output Writer interface
func (z *ZLog) Write(p []byte) (n int, err error) {
	if z.ScreenEnable {
		if z.BuildMode == "rel" {
			n, err = color.New(z.levelColor[string(p[1])]).Println(string(p))
		} else {
			n, err = fmt.Printf("\x1b[%dm"+string(p)+" \x1b[0m\n", z.levelColor[string(p[1])])
		}
	}
	if z.RotateEnable {
		n, err = z.Rotate.Write(p)
	}

	if z.WebsocketEnable {
		n = len(p)
		err = z.WebsocketConn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			z.WebsocketEnable = false
		}
	}
	return
}

// Format implement the Formatter interface
func (z *ZLog) Format(entry *logrus.Entry) ([]byte, error) {
	if z.RotateEnable == false && z.ScreenEnable == false {
		return []byte{}, nil
	}
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	if v, ok := entry.Data["aGVsbG8="]; ok {
		entry.Message = fmt.Sprintf("[%s] %s", v, entry.Message)
		delete(entry.Data, "aGVsbG8=")
	}
	params := MapToJson(entry.Data)
	fileName := ""
	fileLine := 0
	if entry.Caller != nil {
		fileName = path.Base(entry.Caller.File)
		fileLine = entry.Caller.Line
	} else {
		fileName = "pipe.writer"
		entry.Level = logrus.WarnLevel
	}
	b.WriteString(fmt.Sprintf("[%s] %s [%s:%d] %s %s", LevelArray[entry.Level], entry.Time.Format("2006-01-02 15:04:05"), fileName, fileLine, entry.Message, params))
	return b.Bytes(), nil
}

var Log *ZLog

func InitGlobalLogger() *ZLog {
	if Log == nil {
		Log = NewLogger()
	}
	return Log
}

func MapToJson(param map[string]interface{}) string {
	data, _ := json.Marshal(param)
	if bytes.Equal(data, []byte{123, 125}) {
		return ""
	}
	return string(data)
}
