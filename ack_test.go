package main

import "testing"

// TestAcknowledgeClears checks that an acknowledged alarm's banner goes away
// once its condition has ended, even while another alarm is still present,
// and that an alarm present at the acknowledgement stays acknowledged.
func TestAcknowledgeClears(t *testing.T) {
	type step struct {
		sec  float64
		cfg  func(*Config)
		want string // banner key expected afterwards ("" = none)
		ack  bool   // expected acknowledged state of the banner
	}
	run := func(name string, base Config, ackAfter float64, steps []step) {
		now := 1000.0
		p := NewPatient(base, now, 5)
		var a ackState
		adv := func(sec float64) {
			for end := now + sec; now < end; {
				now += 0.25
				p.Update(now)
				a.update(p, now)
			}
		}
		banner := func() (string, bool) {
			if a.on {
				return a.key, true
			}
			return p.AlarmKey, false
		}
		adv(ackAfter)
		if p.AlarmText == "" {
			t.Fatalf("%s: no alarm to acknowledge", name)
		}
		p.Acknowledge()
		a = acknowledge(p, now)
		for i, s := range steps {
			if s.cfg != nil {
				c := p.Cfg
				s.cfg(&c)
				p.Apply(c, now)
			}
			adv(s.sec)
			k, ack := banner()
			if k != s.want || (k != "" && ack != s.ack) {
				t.Errorf("%s, step %d: banner %q (acknowledged %v), want %q (acknowledged %v); active %v",
					name, i, k, ack, s.want, s.ack, p.AlarmKeys)
			}
		}
	}

	// red VTach acknowledged while SpO2 is also low: when the VT ends, the
	// banner moves on to the (also acknowledged) SpO2 alarm, and clears when
	// the SpO2 recovers
	run("VTach + SpO2 low",
		Config{Rhythm: RVTach, Rate: 180, RR: 16, SpO2On: SpO2Monitored, SpO2: 91}, 20,
		[]step{
			{2, nil, "*** VTach", true},
			{15, func(c *Config) { c.Rhythm, c.Rate = RSinus, 80 }, "** SpO₂ Low", true},
			{25, func(c *Config) { c.SpO2 = 98 }, "", false},
		})

	// a one-star arrhythmia alarm acknowledged alone clears once it stops
	run("Vent Bigeminy",
		Config{Rhythm: RSinus, Rate: 80, RR: 16, Ectopy: EctPVCUni, Pattern: PatBigeminy}, 40,
		[]step{
			{25, func(c *Config) { c.Ectopy = EctNone }, "", false},
		})

	// a new alarm after the acknowledgement alarms normally once the
	// acknowledged one has ended
	run("new alarm after ack",
		Config{Rhythm: RVTach, Rate: 180, RR: 16, SpO2On: SpO2Monitored, SpO2: 98}, 20,
		[]step{
			{15, func(c *Config) { c.Rhythm, c.Rate, c.SpO2 = RSinus, 80, 91 }, "** SpO₂ Low", false},
		})
}
