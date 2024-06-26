# [Apache ECharts](https://echarts.apache.org/en/index.html) for Jupyter Notebooks with Go ([Examples](https://janpfeifer.github.io/gonb-echarts/))

This package adds Go Notebook support to [Apache ECharts](https://echarts.apache.org/en/index.html)
using [GoNB](https://github.com/janpfeifer/gonb) Jupyter kernel and [github.com/go-echarts/go-echarts](https://github.com/go-echarts/go-echarts).

It defines two methods to display [go-echarts](https://github.com/go-echarts/go-echarts) charts: `Display`
that immediately display the chart, and `DisplayContent` that returns the HTML content needed to generate
the chart -- useful for instance if the chart needs to be laid out inside other HTML content.

See included [examples](https://janpfeifer.github.io/gonb-echarts/) -- Notebook file (without the rendered images) in [`examples.ipynb`](https://github.com/janpfeifer/gonb-echarts/blob/main/examples.ipynb).

## Screenshots:

**Note**: being just screenshots these are not animated -- see [examples](https://janpfeifer.github.io/gonb-echarts/) for animated version.

### Bar Chart (with code)

```go
import (
	"math/rand"
    
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	gonb_echarts "github.com/janpfeifer/gonb-echarts"
)

// generate random data for bar chart
func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

%%
bar := charts.NewBar()
// set some global options like Title/Legend/ToolTip or anything else
bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
    Title:    "My first bar chart generated by go-echarts",
    Subtitle: "It's extremely easy to use, right?",
}))

// Put data into instance
bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
    AddSeries("Category A", generateBarItems()).
    AddSeries("Category B", generateBarItems())

// Display
err := gonb_echarts.Display(bar, "width: 1024px; height:400px; background: white;")
if err != nil {
    fmt.Printf("Error: %+v\n", err)
}
```

![image](https://github.com/janpfeifer/gonb-echarts/assets/7460115/aa404a22-ad80-4e34-9a3b-5db5da94beca)

### Stacked Lines

![image](https://github.com/janpfeifer/gonb-echarts/assets/7460115/964b253b-f1c0-4a10-9a88-5e1a89327233)

### Geo Map example

![image](https://github.com/janpfeifer/gonb-echarts/assets/7460115/780950af-87cf-47e5-837e-ae616eefe3f4)

## Limitations

Because the charts depend on Javascript, they won't display in GitHub.
But exporting notebooks to HTML using [JupyterLab](https://jupyter.og) will correctly include the charts.

## Issues, feature requests, discussions and support

Please, use [github.com/janpfeifer/gonb](https://github.com/janpfeifer/gonb) repository.
