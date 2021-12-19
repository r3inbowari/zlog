package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/r3inbowari/zlog"
	"net/http"
	"time"
)

var l *zlog.ZLog

func main() {
	l = zlog.NewLogger()
	l.SetRotate(true)
	l.SetScreen(true)
	l.Info("panic")

	a := time.Tick(time.Second)

	for range a {
		l.WithTag("Hell").WithField("good", "girl").Info("hello")
	}
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func log(c *gin.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	l.SetWebsocket(conn)
}
