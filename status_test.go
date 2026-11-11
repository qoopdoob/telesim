package main

import "testing"

// TestRhythmStatus checks the rhythm status text the monitor shows for
// regular rhythms with and without P waves and for paced rhythms.
func TestRhythmStatus(t *testing.T) {
	cases := []struct {
		cfg  Config
		want string
	}{
		{Config{Rhythm: RSinus, Rate: 75}, "Sinus Rhythm"},
		{Config{Rhythm: RSinus, Rate: 120}, "Sinus Tach"},
		{Config{Rhythm: RSinus, Rate: 48}, "Sinus Brady"},
		{Config{Rhythm: RJunctional, Rate: 45}, "SV Brady"},
		{Config{Rhythm: RJunctional, Rate: 75}, "SV Rhythm"},
		{Config{Rhythm: RSVT, Rate: 170}, "SV Tachy"},
		{Config{Rhythm: RAflutter, Rate: 300, FlRatio: Fl2to1}, "SV Tachy"},
		{Config{Rhythm: RAflutter, Rate: 300, FlRatio: Fl4to1}, "SV Rhythm"},
		{Config{Rhythm: RAfib, Rate: 100}, "Atrial Fib"},
		{Config{Rhythm: RPaced, Rate: 70}, "Paced Rhythm"},
		{Config{Rhythm: RAPaced, Rate: 70}, "Paced Rhythm"},
		{Config{Rhythm: RAVPaced, Rate: 70}, "Paced Rhythm"},
		{Config{Rhythm: RPaced, Rate: 60, PaceMode: PaceDemand, IntrRate: 75}, "Paced Rhythm"},
	}
	for i, c := range cases {
		c.cfg.RR = 16
		now := 1000.0
		p := NewPatient(c.cfg, now, int64(i+1))
		for end := now + 60; now < end; {
			now += 0.25
			p.Update(now)
		}
		if p.RhythmStatus != c.want {
			t.Errorf("%s at %d: status %q, want %q", Rhythms[c.cfg.Rhythm].Name, c.cfg.Rate, p.RhythmStatus, c.want)
		}
	}
}
