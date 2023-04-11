package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/dryaf/echo-sse/sse"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	renderer := &TemplateRenderer{
		Templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = renderer

	sseHandler := &sse.SSEHandler{
		Interval: 1 * time.Second,
		Renderer: renderer,
		EventFunc: func(c echo.Context) (string, interface{}, error) {
			return "message", map[string]string{"data": time.Now().Format("15:04:05")}, nil
		},
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.GET("/sse", sseHandler.Handle)

	log.Fatal(e.Start(":8080"))
}
