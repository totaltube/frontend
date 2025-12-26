package helpers

import (
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

var minifier *minify.M

func InitMinifier() {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	// m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDocumentTags:        true,
		KeepEndTags:             true,
		KeepConditionalComments: true,
		KeepQuotes:              true,
	})
	minifier = m
}

func MinifyBytes(html []byte) (minified []byte, err error) {
	minified, err = minifier.Bytes("text/html", html)
	if err != nil {
		return html, err
	}
	return minified, nil
}

func MinifyString(html string) (minified string, err error) {
	minified, err = minifier.String("text/html", html)
	if err != nil {
		return html, err
	}
	return minified, nil
}
