package time

import (
	"testing"
	"time"
)

func TestHumanTimeLower(t *testing.T) {
	str := "3/1/2025"
	datetime, err := ParseTimeLocal(str)
	if err != nil {
		t.Fatal(err)
	}
	lower := HumanTimeLower(datetime, "unknown")
	t.Log(lower)
}

func TestParseTimeOnLocal(t *testing.T) {
	str := "3/1/2014"
	datetime, err := ParseTimeLocal(str)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(datetime)
}

func TestParseTimeIn(t *testing.T) {
	str := "3/1/2014 10:22:22.111"
	datetime, err := ParseTimeIn(str, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(datetime)
}

func TestParseDuration(t *testing.T) {
	str := "0d5h15m40s"
	duration, err := ParseDuration(str)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(duration)
}
