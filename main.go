package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lildude/oura"
	"github.com/lildude/runalyze"
	"golang.org/x/oauth2"
)

var (
	version = "dev"
	appName = fmt.Sprintf("oura-to-runalyze/%s", version)
	start, end string
	yesterday bool
)

func main() {
	godotenv.Load(".env")
	flag.StringVar(&start, "start", "", "Start date in the format: YYYY-MM-DD. If not provided, defaults to Oura's default of one week ago.")
	flag.StringVar(&end, "end", "", "End date in the form: YYYY-MM-DD. If not provided, defaults to Oura's default of today.")
	flag.BoolVar(&yesterday, "yesterday", false, "Use yesterday's date as the start date.")
	flag.Parse()

	if yesterday {
		start = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	sleeps, err := getOuraSleep(start, end)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem getting sleep from Oura: %v\n", err)
		os.Exit(1)
	}

	var sleep []runalyze.Sleep
	var hrRest []runalyze.HeartRateRest

	for _, s := range sleeps {
		sleep = append(sleep, runalyze.Sleep{
			DateTime:           s.BedtimeStart,
			Duration:           secToMin(s.Total),
			RemDuration:        secToMin(s.Rem),
			LightSleepDuration: secToMin(s.Light),
			DeepSleepDuration:  secToMin(s.Deep),
			AwakeDuration:      secToMin(s.Awake),
			Quality:            int(s.Score / 10),
		})

		hrRest = append(hrRest, runalyze.HeartRateRest{
			DateTime:  s.BedtimeEnd,
			HeartRate: s.HrLowest,
		})
	}

	metrics := runalyze.Metrics{
		HeartRateRest: hrRest,
		Sleep:         sleep,
	}

	if err = upLoadMetricsToRunAlyze(metrics); err != nil {
		fmt.Fprintf(os.Stderr, "Problem uploading metrics to Runalyze: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Successfully sync'd to Runalyze")
}

func getOuraSleep(start string, end string) ([]oura.Sleep, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("OURA_ACCESS_TOKEN")})
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	cl := oura.NewClient(tc, appName)
	sleep, _, err := cl.GetSleep(ctx, start, end)

	return sleep.Sleeps, err
}

func upLoadMetricsToRunAlyze(m runalyze.Metrics) error {
	ctx := context.Background()
	cfg := runalyze.Configuration{
		Token: os.Getenv("RUNALYZE_ACCESS_TOKEN"),
		AppName: appName,
	}
	cl := runalyze.NewClient(cfg)
	_, err := cl.CreateMetrics(ctx, m)
	return err
}

func secToMin(sec int) int {
	min, err := time.ParseDuration(fmt.Sprintf("%ds", sec))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return int(min.Minutes() + 0.5)
}
