package main

import (
	"fmt"
	"testing"
)

func Test_limitFileName(t *testing.T) {
	tests := []struct {
		fileName string
		limit    int
		want     string
	}{
		{"абвгрежзик", 15, "абвгрежзик"}, // 1
		{"абвгрежзик", 10, "абвгрежзик"}, // 2
		{"абвгрежзик", 5, "аб…ик"},       // 3
		{"абвгрежзикл", 5, "аб…кл"},      // 4
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i+1), func(t *testing.T) {
			if got := limitFileName(tt.fileName, tt.limit); got != tt.want {
				t.Errorf("limitFileName(%v, %v) = %v, want %v", tt.fileName, tt.limit, got, tt.want)
			}
		})
	}
}
