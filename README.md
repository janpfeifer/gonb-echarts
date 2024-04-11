# [Apache ECharts](https://echarts.apache.org/en/index.html) for Jupyter Notebooks with Go 

This package adds Go Notebook support to [Apache ECharts](https://echarts.apache.org/en/index.html)
using [GoNB](https://github.com/janpfeifer/gonb) Jupyter kernel and [github.com/go-echarts/go-echarts](https://github.com/go-echarts/go-echarts).

## Examples:

*Note*: These are just frozen screen captures. If you open the [Examples Notebook](https://github.com/janpfeifer/gonb-echarts/blob/main/examples.ipynb) in Jupyter Notebook, mouse over will interact with the charts. Unfortunately, GitHub won't display the plots in the [notebook itself](https://github.com/janpfeifer/gonb-echarts/blob/main/examples.ipynb) because it won't
execute javascript in the notebooks.

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

```go
func toLineData[In any](data []In) []opts.LineData {
    r := make([]opts.LineData, len(data))
    for ii, v := range data {
        r[ii].Value = v 
    }
    return r
}

%%
stackedLine := charts.NewLine()
stackedLine.SetGlobalOptions(
    charts.WithTitleOpts(opts.Title{Title: "Stacked Line",}), 
    charts.WithTooltipOpts(opts.Tooltip{Show:true, Trigger: "axis"}),
)
seriesOpt := charts.WithLineChartOpts(opts.LineChart{
    Stack: "Total",
    ShowSymbol: true,
})

stackedLine.
    SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{Type: "value"}))
stackedLine.
    SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}). 
    AddSeries("Email", toLineData([]int{120, 132, 101, 134, 90, 230, 210}), seriesOpt).
    AddSeries("Union Ads", toLineData([]int{220, 182, 191, 234, 290, 330, 310}), seriesOpt).
    AddSeries("Video Ads", toLineData([]int{150, 232, 201, 154, 190, 330, 410}), seriesOpt).
    AddSeries("Direct", toLineData([]int{320, 332, 301, 334, 390, 330, 320}), seriesOpt).
    AddSeries("Search Engine", toLineData([]int{820, 932, 901, 934, 1290, 1330, 1320}), seriesOpt)

must.M(gonb_echarts.Display(stackedLine, "width: 1024px; height:400px; background: white;"))
```

![image](https://github.com/janpfeifer/gonb-echarts/assets/7460115/964b253b-f1c0-4a10-9a88-5e1a89327233)

## Limitations

Because the charts depends on Javascript, they won't display in Github.
But exporting notebooks to HTML using [JupyterLab](https://jupyter.og) will correctly include the charts.

## Issues, feature requests, discussions and support

Please, use [github.com/janpfeifer/gonb](https://github.com/janpfeifer/gonb) repository.
