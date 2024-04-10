// Package echarts provides a convenient rendering of [go-echarts](https://github.com/go-echarts/go-echarts)
// charts and plots for GoNB. It is a wrapper for [Apache ECharts](https://echarts.apache.org/en/index.html).
package echarts

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/janpfeifer/gonb/gonbui"
	"github.com/janpfeifer/gonb/gonbui/dom"
	"github.com/pkg/errors"
	"io"
	"path"
	"path/filepath"
	"strings"
)

// Renderer interface for echarts that implement the `Render` method.
type Renderer interface {
	Render(w io.Writer) error
}

type renderData struct {
	// chartId should be used by the container div that will hold the chart.
	chartId string

	// Script sources.
	jsAssetsSrc []string

	// Javascript code for the specific chart
	jsAssetsCode []string
}

// parseRendering renders given the chart and extract the information needed to re-render it in GoNB.
//
// This is implemented by rendering it to an HTML page (with `<head>` and `<body>` tags) that is then
// parsed
func parseRendering(chart *charts.BaseConfiguration) (data renderData, err error) {
	data.chartId = chart.ChartID
	var buffer bytes.Buffer
	err = chart.Render(&buffer)
	if err != nil {
		err = errors.Wrapf(err, "failed to render chart to a page -- phase one of rendering to notebook")
		return
	}

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(&buffer)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse rendered HTML")
		return
	}

	// Find javascript needed to run the chart.
	var ()
	doc.Find("script").Each(func(i int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if !exists {
			jsCode := selection.Text()
			if jsCode != "" {
				data.jsAssetsCode = append(data.jsAssetsCode, jsCode)
			}
		} else {
			data.jsAssetsSrc = append(data.jsAssetsSrc, src)
		}
	})
	_ = doc
	return
}

type SupportedCharts interface {
	charts.Bar
}

// moduleName tries to guess the module name from a javascript source.
func moduleName(src string) string {
	module := path.Base(src)
	for {
		newModule := strings.TrimSuffix(module, filepath.Ext(module))
		if newModule == module {
			break
		}
		module = newModule
	}
	return module
}

// Display displays the EChart in GoNB.
// The parameter `style` is used for the `<div>` tag that holds the plot. Typically, one will want to set the
// `width` and `height`. E.g.: `style="width: 1024px; height:600px; background: white;"`.
func Display[T SupportedCharts](chart *T, style string) error {
	var data renderData
	var err error
	cAny := any(chart)
	switch c := cAny.(type) {
	case *charts.Bar:
		data, err = parseRendering(&c.BaseConfiguration)
	default:
		err = errors.Errorf("unsupported EChart type %T", cAny)
	}
	if err != nil {
		return err
	}
	if len(data.jsAssetsSrc) == 0 || len(data.jsAssetsCode) == 0 {
		return errors.New("failed to parse javascript of go-echarts rendering")
	}

	// Create containing DIV
	gonbui.DisplayHtmlf(`<div id="%s" style="%s"></div>`, data.chartId, style)

	// Inject needed javascript code. The preamble sets `echarts` from the RequireJS module.
	preamble := `
	if (typeof echarts === 'undefined') {
		window.echarts = module;
	}`
	code := strings.Join(append([]string{preamble}, data.jsAssetsCode...), "\n")
	// Include the first n-1 script sources.
	var noAttr map[string]string
	for ii := 0; ii < len(data.jsAssetsSrc)-1; ii++ {
		src := data.jsAssetsSrc[ii]
		if err = dom.LoadScriptOrRequireJSModuleAndRun(moduleName(src), src, noAttr, ""); err != nil {
			return err
		}
	}
	lastSrc := data.jsAssetsSrc[len(data.jsAssetsSrc)-1]
	if err = dom.LoadScriptOrRequireJSModuleAndRun(moduleName(lastSrc), lastSrc, noAttr, code); err != nil {
		return err
	}
	return nil
}
