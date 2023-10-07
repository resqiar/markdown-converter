package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync"

	chroma_html "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/labstack/echo/v4"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	html "github.com/yuin/goldmark/renderer/html"
)

var (
	md_engine = goldmark.New(
		goldmark.WithExtensions(
			// enable Github Flavoured Markdown
			extension.GFM,
			highlighting.NewHighlighting(
				// set theme for code highlighter.
				// see more theme here: https://swapoff.org/chroma/playground
				highlighting.WithStyle("paraiso-dark"),
				highlighting.WithFormatOptions(
					// enable numbering in the left of the code
					chroma_html.WithLineNumbers(true),
				),
			),
		),
		goldmark.WithParserOptions(
			// enable auto heading for heading,
			// it will automatically generate something like:
			// id="title-of-the-heading", id="i-will-buy-popcorn"
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			// render newlines as <br/>
			html.WithHardWraps(),
		),
	)
	buffer_pool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func main() {
	app := echo.New()

	app.POST("/", func(c echo.Context) error {
		var payload []byte

		// read the body and bind it into payload
		payload, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		// now we call the convert function we created
		// with payload as an argument.
		parsed, err := convert(payload)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		// send the parsed content into a blob of HTML
		return c.Blob(http.StatusOK, "application/html", parsed)
	})

	app.Logger.Fatal(app.Start(":3131"))
}

func convert(s []byte) ([]byte, error) {
	// use buffer from the Pool
	buffer := buffer_pool.Get().(*bytes.Buffer)

	// restore resources back to the pool when done
	defer buffer_pool.Put(buffer)

	// since we are reusing resources, better reset it first
	buffer.Reset()

	// pass markdown and a buffer to Convert
	if err := md_engine.Convert(s, buffer); err != nil {
		log.Println("Error parsing MD:", err)
		return nil, err
	}

	return buffer.Bytes(), nil
}
