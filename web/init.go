package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	// web engine
	router := gin.Default()
	// load html files, tpl files
	router.LoadHTMLGlob("templates/*")
	// load resource files, js/css etc.
	router.Static("/content", "content")
	// set admin router
	setAdminRoute(router)
	// set server info
	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// start and listen port
	s.ListenAndServe()
}
