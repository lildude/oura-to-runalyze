package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lildude/oura"
	"github.com/lildude/runalyze"
)

func Test_secToMin(t *testing.T) {
	tests := []struct {
		name string
		sec  int
		want int
	}{
		{
			name: "120s to 2mins",
			sec:  120,
			want: 2,
		},
		{
			name: "125s to 2mins",
			sec:  120,
			want: 2,
		},
		{
			name: "15s to 0mins",
			sec:  15,
			want: 0,
		},
		{
			name: "30s to 1mins",
			sec:  30,
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := secToMin(tt.sec); got != tt.want {
				t.Errorf("secToMin() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockOuraClient struct{}

func (m *mockOuraClient) GetSleep(ctx context.Context, start, end string) (*oura.Sleeps, *http.Response, error) {
	return &oura.Sleeps{Sleeps: []oura.Sleep{}}, nil, nil
}

func Test_newOuraClient(t *testing.T) {
	want := oura.NewClient(nil).UserAgent
	got := newOuraClient().UserAgent
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("newOuraClient() mismatch (-want +got):\n%s", diff)
	}
}

func Test_getOuraSleep(t *testing.T) {
	want := &[]oura.Sleep{}
	client := &mockOuraClient{}
	got, _ := getOuraSleep(client, "2020-12-24", "2020-12-25")
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("getOuraSleep() mismatch (-want +got):\n%s", diff)
	}
}

type mockRunalyzeClient struct{}

func (m *mockRunalyzeClient) CreateMetrics(ctx context.Context, metrics runalyze.Metrics) (*http.Response, error) {
	if metrics.Sleep[0].Duration > 1440 {
		return &http.Response{}, runalyze.Error{Message: "validation error"}
	}
	return &http.Response{}, nil
}

func Test_newRunalyzeClient(t *testing.T) {
	want := runalyze.NewClient("12345").UserAgent
	got := newRunalyzeClient().UserAgent
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("newRunalyzeClient() mismatch (-want +got):\n%s", diff)
	}
}

func Test_upLoadMetricsToRunAlyze(t *testing.T) {
	type args struct {
		client  runalyzeClient
		metrics runalyze.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "upload succeeds",
			args:    args{client: &mockRunalyzeClient{}, metrics: runalyze.Metrics{Sleep: []runalyze.Sleep{{Duration: 1439}}}},
			wantErr: false,
		},
		{
			name:    "upload fails",
			args:    args{client: &mockRunalyzeClient{}, metrics: runalyze.Metrics{Sleep: []runalyze.Sleep{{Duration: 1441}}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := upLoadMetricsToRunAlyze(tt.args.client, tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("upLoadMetricsToRunAlyze() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createMetrics(t *testing.T) {
	sleeps := []oura.Sleep{
		{
			BedtimeStart: time.Date(2017, 11, 6, 2, 13, 19, 0, time.UTC),
			BedtimeEnd:   time.Date(2017, 11, 6, 8, 12, 19, 0, time.UTC),
			Total:        20310,
			Rem:          7140,
			Light:        10260,
			Deep:         2910,
			Awake:        1230,
			Score:        85,
			HrLowest:     49,
		},
	}
	want := runalyze.Metrics{
		Sleep: []runalyze.Sleep{
			{
				DateTime:           time.Date(2017, 11, 6, 2, 13, 19, 0, time.UTC),
				Duration:           339,
				RemDuration:        119,
				LightSleepDuration: 171,
				DeepSleepDuration:  49,
				AwakeDuration:      21,
				Quality:            8,
			},
		},
		HeartRateRest: []runalyze.HeartRateRest{
			{
				DateTime:  time.Date(2017, 11, 6, 8, 12, 19, 0, time.UTC),
				HeartRate: 49,
			},
		},
	}
	got := createMetrics(sleeps)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ncreateMetrics() mismatch (-want +got):\n%s", diff)
	}
}
