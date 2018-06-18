package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	// web engine
	g := gin.Default()
	// load html files, tpl files
	g.LoadHTMLGlob("templates/*")
	// load resource files, js/css etc.
	g.Static("/content", "content")
	// set admin router
	setAdminRoute(g)
	// set server info
	s := &http.Server{
		Addr:           ":8080",
		Handler:        g,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// start and listen port
	s.ListenAndServe()
}
