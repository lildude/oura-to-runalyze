package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lildude/oura"
	"github.com/lildude/runalyze"
	"golang.org/x/oauth2"
)

func main() {
	godotenv.Load(".env")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	sleeps, err := getOuraSleep(yesterday, "")
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
	cl := oura.NewClient(tc)
	sleep, _, err := cl.GetSleep(ctx, start, end)

	return sleep.Sleeps, err
}

func upLoadMetricsToRunAlyze(m runalyze.Metrics) error {
	ctx := context.Background()
	cl := runalyze.NewClient(nil, os.Getenv("RUNALYZE_ACCESS_TOKEN"))
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
