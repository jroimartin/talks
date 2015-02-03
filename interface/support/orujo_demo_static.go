package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jroimartin/orujo"
	"github.com/jroimartin/orujo-handlers/basic"
	olog "github.com/jroimartin/orujo-handlers/log"
)

// START OMIT
const logLine = `{{.Req.RemoteAddr}} - {{.Req.Method}} {{.Req.RequestURI}}
{{range  $err := .Errors}}  Err: {{$err}}
{{end}}`

func main() {
	s := orujo.NewServer("localhost:8080")

	logger := log.New(os.Stdout, "[orujo] ", log.LstdFlags)
	logHandler := olog.NewLogHandler(logger, logLine)
	authHandler := basic.NewBasicHandler("Restricted Area", "user", "VerySecure")

	s.RouteDefault(http.NotFoundHandler(), orujo.M(logHandler))
	s.Route(`^/`,
		authHandler,
		http.FileServer(http.Dir(".")), // HL
		orujo.M(logHandler),
	)

	logger.Fatalln(s.ListenAndServe())
}

// STOP OMIT
