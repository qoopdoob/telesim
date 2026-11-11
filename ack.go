package main

// Acknowledged alarms at the central station.
//
// Acknowledging (double-click or Silence All) silences every alarm the
// patient has at that moment and freezes the message shown, with a check
// mark. When the frozen alarm's own condition has been gone for 5 s (values
// hovering at a limit should not re-arm it), the station moves on: to the
// next alarm still present if that one was acknowledged too (shown with a
// check mark, silent), otherwise the live alarm display returns (a new alarm
// that arrived after the acknowledgement alarms normally). A higher-priority
// alarm, or an arrhythmia reminder after its timeout, re-arms at once.

// ackState is an acknowledged alarm: the message and level frozen at the
// moment it was silenced, and every alarm that was present then.
type ackState struct {
	on         bool
	key, text  string
	lvl, rank  int
	clearSince float64         // when the frozen alarm's condition went away (0 = still present)
	at         float64         // when it was acknowledged
	keys       map[string]bool // all alarms present when acknowledged
}

// acknowledge freezes the patient's current alarm and remembers every alarm
// present now.
func acknowledge(p *Patient, t float64) ackState {
	a := ackState{on: true, key: p.AlarmKey, text: p.AlarmText, lvl: p.AlarmLvl, rank: AlarmRank(p.AlarmLvl), at: t, keys: map[string]bool{}}
	for _, k := range p.AlarmKeys {
		a.keys[k] = true
	}
	return a
}

// update advances an acknowledgement for the patient's current alarms.
func (a *ackState) update(p *Patient, t float64) {
	if !a.on {
		return
	}
	if AlarmRank(p.AlarmLvl) > a.rank || p.RemindAt > a.at {
		*a = ackState{} // escalation or a reminder: alarm again
		return
	}
	present := false
	for _, k := range p.AlarmKeys {
		if k == a.key {
			present = true
		}
	}
	if present {
		a.clearSince = 0
		return
	}
	if a.clearSince == 0 {
		a.clearSince = t
		return
	}
	if t-a.clearSince < 5 {
		return
	}
	// the frozen alarm has ended: move on to the next acknowledged alarm
	// still present, or back to the live display
	if p.AlarmKey != "" && a.keys[p.AlarmKey] {
		a.key, a.text, a.lvl, a.rank, a.clearSince = p.AlarmKey, p.AlarmText, p.AlarmLvl, AlarmRank(p.AlarmLvl), 0
		return
	}
	*a = ackState{}
}
