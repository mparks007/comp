package circuit

import "errors"

// Reset-Set Flip-Flop

// r s   q  !q
// 0 1   1   0
// 1 0   0   1
// 0 0   q  !q  (hold)
// 1 1   x   x  (invalid)

type rsFlipFlop struct {
	rIn          emitter
	sIn          emitter
	rNor         *norGate
	sNor         *norGate
	sNorTempPin2 emitter // to avoid a direct link between rNor's output as sNor's input (recursion!)
}

func newRSFlipFLop(rPin, sPin emitter) (*rsFlipFlop, error) {
	f := &rsFlipFlop{}

	if err := f.updateInputs(rPin, sPin); err != nil {
		return nil, err
	}

	f.sNorTempPin2 = nil // due to recursion, nil'ing the link from rNor's output.  qEmitting() will handle the rest.

	f.setupNors()

	return f, nil
}

func (f *rsFlipFlop) updateInputs(rPin, sPin emitter) error {
	if err := f.validateInputs(rPin, sPin); err != nil {
		return err
	}

	f.rIn = rPin
	f.sIn = sPin

	return nil
}

func (f *rsFlipFlop) validateInputs(rPin, sPin emitter) error {
	if (rPin != nil && rPin.Emitting()) && (sPin != nil && sPin.Emitting()) {
		return errors.New("Both inputs of a Flip-Flop cannot be powered simultaneously")
	}

	return nil
}

func (f *rsFlipFlop) setupNors() {
	f.sNor = newNORGate(f.sIn, f.sNorTempPin2)
	f.rNor = newNORGate(f.rIn, f.sNor)
}

func (f *rsFlipFlop) qEmitting() (bool, error) {

	if err := f.validateInputs(f.rIn, f.sIn); err != nil {
		return false, err
	}

	f.setupNors()

	if f.rNor.Emitting() {
		f.sNorTempPin2 = &Battery{}
		return true, nil
	} else {
		f.sNorTempPin2 = nil
		return false, nil
	}
}

func (f *rsFlipFlop) qBarEmitting() (bool, error) {
	if qEmitting, err := f.qEmitting(); err != nil {
		return !qEmitting, err
	} else {
		return !qEmitting, nil
	}
}
