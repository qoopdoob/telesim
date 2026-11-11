package main

import (
	"math"
	"strings"
	"testing"
)

// TestArrestBreathing checks that breathing stops in cardiac arrest (after a
// short period of agonal gasps) and starts again after a pulse returns, for
// asystole, VF and pulseless VT.
func TestArrestBreathing(t *testing.T) {
	arrests := []struct {
		name string
		cfg  func(Config) Config
	}{
		{"asystole", func(c Config) Config { c.Rhythm = RAsystole; return c }},
		{"VF", func(c Config) Config { c.Rhythm = RVFibCoarse; c.Rate = 0; return c }},
		{"pulseless VT", func(c Config) Config { c.Rhythm = RVTach; c.Rate = 180; c.Pulseless = true; return c }},
	}
	for seed, a := range arrests {
		for run := 0; run < 4; run++ {
			base := Config{Rhythm: RSinus, Rate: 80, RR: 16, SpO2On: SpO2Monitored, SpO2: 97, CO2On: true, EtCO2: 38, RespWaveOn: true}
			now := 1000.0
			p := NewPatient(base, now, int64(100*seed+run))
			step := func(until float64, f func(t float64)) {
				for now < until {
					now += 0.25
					p.Update(now)
					if f != nil {
						f(now)
					}
				}
			}
			step(1060, nil)
			if p.RRDisp < 12 || p.ApneaNow {
				t.Fatalf("%s: before the arrest RR=%d apnea=%v", a.name, p.RRDisp, p.ApneaNow)
			}
			arrestAt := now
			p.Apply(a.cfg(base), now)
			var apneaAt float64
			lastBreath := arrestAt
			gasps := 0
			minRR := 99
			step(arrestAt+240, func(t float64) {
				if p.ApneaNow && apneaAt == 0 {
					apneaAt = t
				}
				if n := len(p.breaths); n > 0 && p.breaths[n-1] > lastBreath && p.breaths[n-1] <= t {
					lastBreath = p.breaths[n-1]
					if lastBreath > arrestAt+16 {
						gasps++
					}
				}
				if !p.ApneaNow && p.RRDisp > 0 && p.RRDisp < minRR && t > arrestAt+20 {
					minRR = p.RRDisp
				}
			})
			if apneaAt == 0 {
				t.Fatalf("%s: no apnea within 4 min of the arrest", a.name)
			}
			if d := apneaAt - arrestAt; d < 45 || d > 135 {
				t.Errorf("%s: apnea %.0f s after the arrest, want about 1-2 min", a.name, d)
			}
			if now-lastBreath < 100 {
				t.Errorf("%s: still breathing %.0f s into the arrest", a.name, lastBreath-arrestAt)
			}
			if gasps < 2 {
				t.Errorf("%s: only %d agonal gasps", a.name, gasps)
			}
			if run == 0 {
				t.Logf("%-13s apnea alarm after %3.0f s, last gasp at %3.0f s, %d gasps, lowest RR shown %d/min, etCO2 %d",
					a.name, apneaAt-arrestAt, lastBreath-arrestAt, gasps, minRR, p.EtCO2Disp)
			}
			found := false
			for _, h := range p.History {
				if strings.HasPrefix(h.Key, "*** Apnea") {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: no *** Apnea alarm in the review", a.name)
			}

			// return of spontaneous circulation
			rosc := now
			p.Apply(base, now)
			resumed := math.Inf(1)
			step(rosc+90, func(t float64) {
				if !p.ApneaNow && math.IsInf(resumed, 1) {
					resumed = t
				}
			})
			if d := resumed - rosc; d < 5 || d > 25 {
				t.Errorf("%s: breathing resumed %.0f s after ROSC", a.name, d)
			}
			if p.RRDisp < 12 || p.RRDisp > 20 {
				t.Errorf("%s: RR %d a minute and a half after ROSC", a.name, p.RRDisp)
			}
			if run == 0 {
				t.Logf("%-13s after ROSC: apnea cleared after %.0f s, RR %d, etCO2 %d", a.name, resumed-rosc, p.RRDisp, p.EtCO2Disp)
			}
		}
	}
}
