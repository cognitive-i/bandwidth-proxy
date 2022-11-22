package proxy

import (
	"embed"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

//go:embed html/*
var htmlFs embed.FS
var htmlTemplate = template.Must(template.ParseFS(htmlFs, "html/*"))

type htmlFields struct {
	Error   string
	Bitrate int64
}

type controlServer struct {
	*gin.Engine
	config *config
}

func newControlServer(config *config) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	r := &controlServer{gin.New(), config}
	r.SetHTMLTemplate(htmlTemplate)
	r.GET("/", r.GetPage)
	r.POST("/", r.SetConfig)

	return &http.Server{Handler: r}
}

func (cs *controlServer) GetPage(ctx *gin.Context) {
	fields := htmlFields{
		Bitrate: cs.config.GetMaxBitrate(),
	}

	ctx.HTML(http.StatusOK, "index.html", fields)
}

func (cs *controlServer) SetConfig(ctx *gin.Context) {
	ctx.Request.ParseForm()

	status := http.StatusNotAcceptable
	fields := htmlFields{}
	if bitrate, err := strconv.ParseInt(ctx.Request.Form.Get("bitrate"), 10, 64); err != nil {
		fields.Error = err.Error()
	} else if err := cs.config.SetMaxBitrate(bitrate); err != nil {
		fields.Error = err.Error()
	} else {
		status = http.StatusAccepted
	}

	fields.Bitrate = cs.config.GetMaxBitrate()
	ctx.HTML(status, "index.html", fields)
}

type client string

var clientErrorMessage = regexp.MustCompilePOSIX("Error:(.+?)<br />")

func Client(controlAddress string) client {
	u := &url.URL{Scheme: "http", Host: controlAddress}
	return client(u.String())
}

func (c client) SetMaxBitrate(bandwidth int) client {
	v := url.Values{}
	v.Set("bitrate", strconv.Itoa(bandwidth))

	if r, err := http.PostForm(string(c), v); err != nil {
		log.Fatal(err)
	} else if r.StatusCode != http.StatusAccepted {
		details := ""
		if b, err := io.ReadAll(r.Body); err == nil {
			details = string(b)
			if m := clientErrorMessage.FindStringSubmatch(details); len(m) > 1 {
				template.HTMLEscaper()
				details = html.UnescapeString(m[1])
			}
		}
		log.Fatal("SetMaxBitrate failed", details)
	} else {
		log.Println("SetMaxBitrate successful")
	}

	return c
}
