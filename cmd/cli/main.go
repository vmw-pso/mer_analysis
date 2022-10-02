package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/elliotchance/orderedmap"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	mer "github.com/vmw-pso/eve/mer"
)

type config struct {
	MAU          int          `json:"mau"`
	Distribution Distribution `json:"distribution"`
	Systems      Systems      `json:"systems"`
}

type Distribution struct {
	Highsec  float64 `json:"highsec"`
	Lowsec   float64 `json:"lowsec"`
	Nullsec  float64 `json:"nullsec"`
	Wormhole float64 `json:"wormhole"`
}

type Systems struct {
	Highsec  int `json:"highsec"`
	Lowsec   int `json:"lowsec"`
	Nullsec  int `json:"nullsec"`
	Jove     int `json:"jove"`
	Wormhole int `json:"wormhole"`
	Abbysal  int `json:"abyssal"`
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		logging = flags.Bool("log", false, "whether to log to file or not")
		convert = flags.String("convert", "", "filename to convert")
		analyze = flags.String("analyze", "", "file to analyze for monthly statistics")
		plot    = flags.Bool("plot", false, "whether to analyse and plot all data")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c := config{}
	c.loadConfigFromJson("config.json")

	if *logging {
		log.Println("Logging true")
	}

	filename := *convert
	if filename != "" {
		converter := mer.NewConverter(filename)
		return converter.Convert()
	}

	filename = *analyze
	if filename != "" {
		analyzer, err := mer.NewAnalyzer(filename)
		if err != nil {
			return err
		}
		stats, err := analyzer.Analyze()
		if err != nil {
			return err
		}

		fmt.Println("---------------------------------------------------------------------------------------")
		fmt.Println("|System Class  |Total Losses|Systems|Loss/System|Character Distribution|Loss/Character|")
		fmt.Println("---------------------------------------------------------------------------------------")
		fmt.Printf("|Highsec       |%12d|%7d|%11.1f|%22.2f|%14.2f|\n", stats.HighsecKills, c.Systems.Highsec, float64(stats.HighsecKills)/float64(c.Systems.Highsec), c.Distribution.Highsec, float64(stats.HighsecKills)/(float64(c.MAU)*c.Distribution.Highsec))
		fmt.Printf("|Lowsec        |%12d|%7d|%11.1f|%22.2f|%14.2f|\n", stats.LowsecKills, c.Systems.Lowsec, float64(stats.LowsecKills)/float64(c.Systems.Lowsec), c.Distribution.Lowsec, float64(stats.LowsecKills)/(float64(c.MAU)*c.Distribution.Lowsec))
		fmt.Printf("|Nullsec       |%12d|%7d|%11.1f|%22.2f|%14.2f|\n", stats.NullsecKills, c.Systems.Nullsec, float64(stats.NullsecKills)/float64(c.Systems.Nullsec), c.Distribution.Nullsec, float64(stats.NullsecKills)/(float64(c.MAU)*c.Distribution.Nullsec))
		fmt.Printf("|Wormhole      |%12d|%7d|%11.1f|%22.2f|%14.2f|\n", stats.WormholeKills, c.Systems.Wormhole, float64(stats.WormholeKills)/float64(c.Systems.Wormhole), c.Distribution.Wormhole, float64(stats.WormholeKills)/(float64(c.MAU)*c.Distribution.Wormhole))
		fmt.Printf("|Abyss/Pochven |%12d|%7d|%11.1f|%22.2s|%14.2s|\n", stats.AbyssalKills, c.Systems.Abbysal, float64(stats.AbyssalKills)/float64(c.Systems.Abbysal), "", "")
		fmt.Println("---------------------------------------------------------------------------------------")
		fmt.Printf("|Total         |%12d|%7s|%11.1s|%22.2s|%14.2s|\n", stats.TotalKills, "", "", "", "")
		fmt.Println("---------------------------------------------------------------------------------------")
	}

	if *plot {
		cs := orderedmap.NewOrderedMap() // cs = cumulativeStats
		year := 2016
		month := 6

		labels := []string{}
		totalKills := []opts.LineData{}
		highsecKills := []opts.LineData{}
		lowsecKills := []opts.LineData{}
		nullsecKills := []opts.LineData{}
		wormholeKills := []opts.LineData{}
		abyssalKills := []opts.LineData{}

		shouldStop := false

		for !shouldStop {
			monthStr := monthString(year, month)
			filename := fmt.Sprintf("%s_kill_dump.csv", monthStr)
			path := fmt.Sprintf("./assets/killdump/%s", filename)
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				shouldStop = true
			} else {
				labels = append(labels, monthStr)
				fmt.Println(path)
				analyzer, err := mer.NewAnalyzer(path)
				if err != nil {
					return err
				}
				stats, err := analyzer.Analyze()
				if err != nil {
					return err
				}
				totalKills = append(totalKills, opts.LineData{Value: stats.TotalKills})
				highsecKills = append(highsecKills, opts.LineData{Value: stats.HighsecKills})
				lowsecKills = append(lowsecKills, opts.LineData{Value: stats.LowsecKills})
				nullsecKills = append(nullsecKills, opts.LineData{Value: stats.NullsecKills})
				wormholeKills = append(wormholeKills, opts.LineData{Value: stats.WormholeKills})
				abyssalKills = append(abyssalKills, opts.LineData{Value: stats.AbyssalKills})
			}
			month++
			if month == 13 {
				month = 1
				year++
			}
		}
		for el := cs.Front(); el != nil; el = el.Next() {
			fmt.Println(el.Key, el.Value)
		}

		line := charts.NewLine()
		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
			charts.WithTitleOpts(opts.Title{
				Title:    "MER Killmails By Security Class",
				Subtitle: "Source data: Monthly Economic Report (MER) Jun 2016 - Latest",
			}),
			charts.WithLegendOpts(opts.Legend{
				Show: true,
			}),
			//charts.WithColorsOpts(opts.Colors{"#2596be", "#154c79"}),
		)

		line.SetXAxis(labels).
			AddSeries("Total", totalKills).
			AddSeries("Highsec", highsecKills).
			AddSeries("Lowsec", lowsecKills).
			AddSeries("Nullsec", nullsecKills).
			AddSeries("Wormhole", wormholeKills).
			AddSeries("Abyssal", abyssalKills).
			SetSeriesOptions(
				charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
				charts.WithLineStyleOpts(opts.LineStyle{Width: 2.0}),
			)

		filename := "mer_kill_dump.html"
		f, err := os.Create(filename)
		if err != nil {
			fmt.Printf("file error: %v\n", err.Error())
			os.Exit(1)
		}
		line.Render(f)
	}

	return nil
}

func monthString(year, month int) string {
	if countDigits(month) == 1 {
		return fmt.Sprintf("%d0%d", year, month)
	}
	return fmt.Sprintf("%d%d", year, month)
}

func countDigits(number int) int {
	count := 0
	for number > 0 {
		number = number / 10
		count++
	}
	return count
}

func (c *config) loadConfigFromJson(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}

	return nil
}
