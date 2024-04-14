// Package echarts provides a convenient rendering of [go-echarts](https://github.com/go-echarts/go-echarts)
// charts and plots for GoNB. It is a wrapper for [Apache ECharts](https://echarts.apache.org/en/index.html).
//
// It defines two methods to display [go-echarts](https://github.com/go-echarts/go-echarts) charts: `Display`
// that immediately display the chart, and `DisplayContent` that returns the HTML content needed to generate
// the chart -- useful for instance if the chart needs to be laid out inside other HTML content.
//
// See include `examples.ipynb` for examples.
package echarts

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/janpfeifer/gonb/gonbui"
	"github.com/pkg/errors"
	"io"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer interface for echarts that implement the `Render` method.
type Renderer interface {
	Render(w io.Writer) error
}

// renderData parsed from go-echarts rendering, and re-used for GoNB rendering.
type renderData struct {
	// ChartId should be used by the container div that will hold the chart.
	ChartId string

	// Script sources.
	JsAssetsSrc []string

	// JsAssetsCode code for the specific chart
	JsAssetsCode []string
}

// parseRendering renders given the chart and extract the information needed to re-render it in GoNB.
//
// This is implemented by rendering it to an HTML page (with `<head>` and `<body>` tags) that is then
// parsed
func parseRendering(chart *charts.BaseConfiguration) (data renderData, err error) {
	data.ChartId = chart.ChartID
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
				data.JsAssetsCode = append(data.JsAssetsCode, jsCode)
			}
		} else {
			data.JsAssetsSrc = append(data.JsAssetsSrc, src)
		}
	})
	_ = doc
	return
}

type SupportedCharts interface {
	charts.Bar | charts.Bar3D | charts.BoxPlot | charts.Custom | charts.EffectScatter | charts.Funnel |
		charts.Gauge | charts.Geo | charts.Graph | charts.HeatMap | charts.Kline | charts.Line3D | charts.Line | charts.Liquid |
		charts.Map | charts.Parallel | charts.Pie | charts.Radar | charts.Sankey | charts.Scatter3D | charts.Scatter | charts.Sunburst |
		charts.Surface3D | charts.ThemeRiver | charts.Tree | charts.TreeMap | charts.WordCloud
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

var displayTmpl = template.Must(template.New("display").Parse(`
(() => {
	let echartsFn = function() {
	{{range .JsAssetsCode}}
		{{.}}
	{{end}}
	}

	let echartsSrcs = [
	{{range .JsAssetsSrc}}
		"{{.}}",	
	{{end}}
	];

	function loadScriptsThenExecute(scripts, fn) {
		if (scripts.length == 0) {
			// Nothing to load, execute immediately.
			fn();
			return;
		}

		// Keep track of loaded scripts
		let loadedCount = 0;
		const head = document.head;
		
		// Function to handle successful script loading
		const scriptLoaded = () => {
			loadedCount++;
			if (loadedCount === scripts.length) {
				fn(); // Execute the callback function when all scripts are loaded
			}
		};
		
		for (const src of scripts) {
			// Check if script is already loaded
			const existingScript = document.querySelector('script[src="'+src+'"]');
			if (existingScript) {
				// Script already loaded. 
				scriptLoaded(); // Proceed as if loaded
			} else if (typeof requirejs === "function") {
				require([src], function(loadedModule) {
        			// Note: 'loadedModule' will contain the exports from the script, if any.
        			scriptLoaded(); 
				});
			} else {
				// Create the script element
				const script = document.createElement('script');
				script.async = false;  // Order matters, this must be false.
				script.src = src;
				script.onload = scriptLoaded;
				script.onerror = () => console.error('Failed to load script: '+src);
				head.appendChild(script);
			}
		}
	}

	if (typeof requirejs === "function") {
		console.log("Using RequireJS");
		let src = echartsSrcs.shift();  // The first source is echarts, which must be loaded with RequireJS. 
		// Use RequireJS to load module.
		let srcWithoutExtension = src.substring(0, src.lastIndexOf(".js"));
		requirejs.config({
			paths: {
				'echarts': srcWithoutExtension
			}
		});
		require(['echarts'], function(echarts) {
			window.echarts = echarts;  // Define echarts globally.
			loadScriptsThenExecute(echartsSrcs, echartsFn);  // Load rest of scripts.	
		});
		return

	} else {
		console.log("Not using RequireJS");
		loadScriptsThenExecute(echartsSrcs, echartsFn);
	}

})();
`))

// Display displays the EChart in GoNB.
// The parameter `style` is used for the `<div>` tag that holds the plot. Typically, one will want to set the
// `width` and `height`. E.g.: `style="width: 1024px; height:600px; background: white;"`.
func Display[T SupportedCharts](chart *T, style string) error {
	html, err := DisplayContent(chart, style)
	if err != nil {
		return err
	}
	gonbui.DisplayHtml(html)
	return nil
}

// DisplayContent returns the HTML content (including a `<script>` tag) that displays the EChart in GoNB.
// One can used [Display] to display it directly, but if one wants to compose or change the layout, one can use
// this instead.
//
// The parameter `style` is used for the `<div>` tag that holds the plot. Typically, one will want to set the
// `width` and `height`. E.g.: `style="width: 1024px; height:600px; background: white;"`.
func DisplayContent[T SupportedCharts](chart *T, style string) (html string, err error) {
	var data renderData
	cAny := any(chart)
	switch c := cAny.(type) {
	case *charts.Bar:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Bar3D:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.BoxPlot:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Custom:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.EffectScatter:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Funnel:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Gauge:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Geo:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Graph:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.HeatMap:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Kline:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Line3D:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Line:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Liquid:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Map:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Parallel:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Pie:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Radar:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Sankey:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Scatter3D:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Scatter:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Sunburst:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Surface3D:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.ThemeRiver:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.Tree:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.TreeMap:
		data, err = parseRendering(&c.BaseConfiguration)
	case *charts.WordCloud:
		data, err = parseRendering(&c.BaseConfiguration)
	default:
		err = errors.Errorf("unsupported EChart type %T", cAny)
	}
	if err != nil {
		return
	}
	if len(data.JsAssetsSrc) == 0 || len(data.JsAssetsCode) == 0 {
		err = errors.New("failed to parse javascript of go-echarts rendering")
		return
	}

	// Generate code.
	var code bytes.Buffer
	err = displayTmpl.Execute(&code, &data)
	if err != nil {
		err = errors.Wrapf(err, "failed to executed template of javascript code to build the echart")
		return
	}

	// Render HTML.
	html = fmt.Sprintf(`<div id="%s" style="%s"></div><script>%s</script>`, data.ChartId, style, code.String())
	return
}
