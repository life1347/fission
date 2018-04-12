package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/mongodb/mongo-tools/common/json"
	"github.com/wcharczuk/go-chart"
)

var (
	file       *string
	outputFile *string
)

type (
	Tags struct {
		Group  string `json:"group"`
		Iter   string `json:"iter"`
		Method string `json:"method"`
		Name   string `json:"name"`
		Proto  string `json:"proto"`
		Status string `json:"status"`
		URL    string `json:"url"`
		Vu     string `json:"vu"`
	}

	Data struct {
		Time  time.Time `json:"time"`
		Value float64   `json:"value"`
		Tags  Tags      `json:"tags"`
	}

	MetricPoint struct {
		Type   string `json:"type"`
		Data   Data   `json:"data"`
		Metric string `json:"metric"`
	}
)

func init() {
	file = flag.String("file", "", "Metric json file name")
	outputFile = flag.String("o", "chart.png", "Output png file name")
	flag.Parse()
}

func getFileReader(file string) (*bufio.Reader, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(f), nil
}

func generateChart(file string, xVals []float64, yVals []float64) error {
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				},
				XValues: xVals,
				YValues: yVals,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, buffer.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	if file == nil || len(*file) == 0 {
		fmt.Println("Please provide metric json file name")
		return
	}

	if len(*outputFile) == 0 {
		fmt.Println("Please output chart png file name")
		return
	}

	reader, err := getFileReader(*file)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	var points []*MetricPoint
	var xVals []float64
	var yVals []float64

	var initTime *time.Time

	for {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		point := &MetricPoint{}
		err = json.Unmarshal(l, point)
		if err != nil {
			fmt.Printf("Failed to parse metric point: %v", err)
			return
		}
		points = append(points, point)

		if initTime == nil {
			initTime = &point.Data.Time
		}

		timeSinceStart := point.Data.Time.Sub(*initTime).Seconds()
		xVals = append(xVals, timeSinceStart)
		yVals = append(yVals, point.Data.Value)
	}

	err = generateChart(*outputFile, xVals, yVals)
	if err != nil {
		fmt.Printf("Failed to generate chart: %v\n", err)
	}
}
