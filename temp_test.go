package main

import (
	"math"
	"strings"
	"testing"
)

// TestTemperature checks the temperature probe: the reading follows the
// patient's temperature slowly, and fever / low temperature give the
// ** Temp High / ** Temp Low alarms with the value in the display unit.
func TestTemperature(t *testing.T) {
	defer func() { TempF = true }()
	base := Config{Rhythm: RSinus, Rate: 80, RR: 16, TempOn: true, Temp: 37.0}
	now := 1000.0
	p := NewPatient(base, now, 1)
	run := func(sec float64) {
		for end := now + sec; now < end; {
			now += 0.25
			p.Update(now)
		}
	}
	run(30)
	if !p.TempValid || math.Abs(p.TempDisp-37) > 0.2 {
		t.Fatalf("normal temperature reads %.2f (valid %v)", p.TempDisp, p.TempValid)
	}
	if p.AlarmText != "" {
		t.Fatalf("alarm at a normal temperature: %q", p.AlarmText)
	}

	// fever: the reading climbs over tens of seconds, then alarms
	c := base
	c.Temp = 39.0
	p.Apply(c, now)
	run(5)
	if p.TempDisp > 38.2 {
		t.Errorf("temperature jumped to %.2f within 5 s", p.TempDisp)
	}
	run(120)
	if math.Abs(p.TempDisp-39) > 0.2 {
		t.Errorf("fever reads %.2f after 2 min, want ~39.0", p.TempDisp)
	}
	TempF = true
	run(1)
	if !strings.HasPrefix(p.AlarmText, "** Temp 102.") || !strings.HasSuffix(p.AlarmText, ">100.4") {
		t.Errorf("fever alarm text %q, want ** Temp 102.x>100.4", p.AlarmText)
	}
	TempF = false
	run(1)
	if !strings.HasPrefix(p.AlarmText, "** Temp 39.") || !strings.HasSuffix(p.AlarmText, ">38.0") {
		t.Errorf("fever alarm text in °C %q, want ** Temp 39.x>38.0", p.AlarmText)
	}
	TempF = true

	// low temperature
	c.Temp = 35.2
	p.Apply(c, now)
	run(150)
	if p.AlarmKey != "** Temp Low" {
		t.Errorf("low temperature: alarm %q, want ** Temp Low", p.AlarmKey)
	}

	// probe off: no reading, no alarm
	c.TempOn = false
	p.Apply(c, now)
	run(10)
	if p.TempValid || strings.Contains(p.AlarmText, "Temp") {
		t.Errorf("probe off: valid %v, alarm %q", p.TempValid, p.AlarmText)
	}

	// unit conversion round trip
	for _, f := range []bool{true, false} {
		TempF = f
		for _, x := range []float64{35, 36.6, 38, 40.1} {
			if math.Abs(TempIn(TempOut(x))-x) > 1e-9 {
				t.Errorf("unit round trip %v -> %v", x, TempIn(TempOut(x)))
			}
		}
	}
	TempF = true
	if TempStr(37) != "98.6" || TempStr(38) != "100.4" {
		t.Errorf("TempStr: 37 °C = %s °F, 38 °C = %s °F", TempStr(37), TempStr(38))
	}
}
