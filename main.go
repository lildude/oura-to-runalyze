package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lildude/oura"
	"github.com/lildude/runalyze"
	"golang.org/x/oauth2"
)

var (
	version            = "dev"
	commit, start, end string
	yesterday          bool
)

func main() {
	godotenv.Load(".env")
	flag.StringVar(&start, "start", "", "Start date in the format: YYYY-MM-DD. If not provided, defaults to Oura's default of one week ago.")
	flag.StringVar(&end, "end", "", "End date in the form: YYYY-MM-DD. If not provided, defaults to Oura's default of today.")
	flag.BoolVar(&yesterday, "yesterday", false, "Use yesterday's date as the start date.")
	ver := flag.Bool("version", false, "Print the version and exit.")
	flag.Parse()

	if *ver {
		fmt.Printf("oura-to-runalyze %s (commit: %s)", version, commit)
		os.Exit(0)
	}

	if _, ok := os.LookupEnv("OURA_ACCESS_TOKEN"); !ok {
		fmt.Printf("%s not set\n", "OURA_ACCESS_TOKEN")
		os.Exit(1)
	}

	if _, ok := os.LookupEnv("RUNALYZE_ACCESS_TOKEN"); !ok {
		fmt.Printf("%s not set\n", "RUNALYZE_ACCESS_TOKEN")
		os.Exit(1)
	}

	if yesterday {
		start = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	ouraClient := newOuraClient()
	sleeps, err := getOuraSleep(ouraClient, start, end)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem uploading metrics to Runalyze: %v\n", err)
		os.Exit(1)
	}
	metrics := createMetrics(*sleeps)

	runalyzeClient := newRunalyzeClient()
	if err := upLoadMetricsToRunAlyze(runalyzeClient, metrics); err != nil {
		fmt.Fprintf(os.Stderr, "Problem uploading metrics to Runalyze: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully sync'd to Runalyze")
}

func createMetrics(sleeps []oura.Sleep) runalyze.Metrics {
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

	return metrics
}

type ouraClient interface {
	GetSleep(ctx context.Context, start, end string) (*oura.Sleeps, *http.Response, error)
}

func newOuraClient() *oura.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("OURA_ACCESS_TOKEN")})
	tc := oauth2.NewClient(context.Background(), ts)
	cl := *oura.NewClient(tc)
	return &cl
}

func getOuraSleep(client ouraClient, start string, end string) (*[]oura.Sleep, error) {
	sleep, _, err := client.GetSleep(context.Background(), start, end)
	return &sleep.Sleeps, err
}

type runalyzeClient interface {
	CreateMetrics(ctx context.Context, metrics runalyze.Metrics) (*http.Response, error)
}

func newRunalyzeClient() *runalyze.Client {
	cl := *runalyze.NewClient(os.Getenv("RUNALYZE_ACCESS_TOKEN"))
	return &cl
}

func upLoadMetricsToRunAlyze(client runalyzeClient, metrics runalyze.Metrics) error {
	_, err := client.CreateMetrics(context.Background(), metrics)
	return err
}

func secToMin(sec int) int {
	min, _ := time.ParseDuration(fmt.Sprintf("%ds", sec))
	return int(min.Minutes() + 0.5)
}
