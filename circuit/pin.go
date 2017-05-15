package circuit

// Basic pin with a single power source that is either powered or not
type inPin struct {
	pwrSourceA emitter
}

func newInPin(a emitter) *inPin {
	return &inPin{a}
}

func (in *inPin) Emitting() bool {
	return in.pwrSourceA != nil && in.pwrSourceA.Emitting()
}

// Special output pin that is only receiving power if the relay is disengaged/off (open switch)
type outOpenPin struct {
	pwrSourceA emitter
	pwrSourceB emitter
}

func newOutOpenPin(a emitter, b emitter) *outOpenPin {
	return &outOpenPin{a, b}
}

func (out *outOpenPin) Emitting() bool {
	if out == nil {
		return false
	}

	return out.pwrSourceA != nil && out.pwrSourceA.Emitting() &&
		(out.pwrSourceB == nil ||
			(out.pwrSourceB != nil && !out.pwrSourceB.Emitting()))
}

// Special output pin that is only receiving power if the relay is engaged/on (closed switch)
type outClosedPin struct {
	pwrSourceA emitter
	pwrSourceB emitter
}

func newOutClosedPin(a emitter, b emitter) *outClosedPin {
	return &outClosedPin{a, b}
}

func (out *outClosedPin) Emitting() bool {
	if out == nil {
		return false
	}
	return out.pwrSourceA != nil && out.pwrSourceA.Emitting() &&
		out.pwrSourceB != nil && out.pwrSourceB.Emitting()
}
