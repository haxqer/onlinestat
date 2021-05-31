package main

import "testing"

func TestGetTodayBeginning(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{name: "test1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTodayBeginning(); got != tt.want {
				t.Errorf("GetTodayBeginning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTomorrowBeginning(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{name: "test1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTomorrowBeginning(); got != tt.want {
				t.Errorf("GetTomorrowBeginning() = %v, want %v", got, tt.want)
			}
		})
	}
}