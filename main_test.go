package main

import (
	"testing"
	"time"

	"github.com/fluent/fluent-bit-go/output"
)

func TestParseRoomIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    []int64
		wantErr bool
	}{
		{name: "empty", input: "", wantErr: true},
		{name: "single", input: "123", want: []int64{123}},
		{name: "spaced", input: " 1 , 2 , 3 ", want: []int64{1, 2, 3}},
		{name: "invalid", input: "1,x", wantErr: true},
		{name: "only_commas", input: " , , ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseRoomIDs(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRoomIDs: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("idx %d: got %d want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseRecordTime(t *testing.T) {
	t.Parallel()
	flbT := output.FLBTime{Time: time.Unix(100, 0).UTC()}
	got := parseRecordTime(flbT)
	if !got.Equal(flbT.Time) {
		t.Fatalf("FLBTime: got %v want %v", got, flbT.Time)
	}
	got = parseRecordTime(uint64(200))
	want := time.Unix(200, 0)
	if !got.Equal(want) {
		t.Fatalf("uint64: got %v want %v", got, want)
	}
}

func TestFormatRecordValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		floorFloat bool
		val        interface{}
		want       string
	}{
		{name: "string", val: "a", want: "a"},
		{name: "bytes", val: []byte("b"), want: "b"},
		{name: "float full", floorFloat: false, val: float64(1.25), want: "1.250000"},
		{name: "float round", floorFloat: true, val: float64(1.6), want: "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatRecordValue(tt.floorFloat, tt.val)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestBuildOptionalLines(t *testing.T) {
	t.Parallel()
	m := map[string]string{"a": "1", "b": "2"}
	got := buildOptionalLines(m, []string{"b", "missing", "a"})
	if got != "- b: 2\n- a: 1\n" {
		t.Fatalf("unexpected lines:\n%q", got)
	}
}
