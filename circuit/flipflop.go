package circuit

import "errors"

type rsFlipFlop struct {
	rIn          emitter
	sIn          emitter
	rNor         *norGate
	sNor         *norGate
	sNorTempPin2 emitter
}

func newRSFlipFLop(rPin, sPin emitter) *rsFlipFlop {
	f := &rsFlipFlop{}

	f.sNorTempPin2 = nil // due to recursion, nil'ing the link to rNor's output.  qEmitting() will handle this.

	f.updateInputs(rPin, sPin)
	f.setupNors()

	return f
}

func (f *rsFlipFlop) updateInputs(rPin, sPin emitter) error {
	if (rPin != nil && rPin.Emitting()) && (sPin != nil && sPin.Emitting()) {
		return errors.New("Both inputs of an RS FlipFlop cannot be powered simultaneously.")
	}

	f.rIn = rPin
	f.sIn = sPin

	return nil
}

func (f *rsFlipFlop) setupNors() {
	f.sNor = newNORGate(f.sIn, f.sNorTempPin2)
	f.rNor = newNORGate(f.rIn, f.sNor)
}

// r s   q  !q
// 0 1   1   0
// 1 0   0   1
// 0 0   q  !q  (hold)
// 1 1   x   x  (invalid)

func (f *rsFlipFlop) qEmitting() bool {

	f.setupNors()

	if f.rNor.Emitting() {
		f.sNorTempPin2 = &Battery{}
		return true
	} else {
		f.sNorTempPin2 = nil
		return false
	}
}

func (f *rsFlipFlop) qBarEmitting() bool {
	return !f.qEmitting()
}
