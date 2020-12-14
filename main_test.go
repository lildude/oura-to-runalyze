package main

import (
	"testing"
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
