package circuit

import "fmt"

// Level-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk    q  !q
// 0  1     0  1
// 1  1     1  0
// X  0     q  !q  (data doesn't matter, no clock high to trigger a store-it action)

type LevelTriggeredDTypeLatch struct {
	rs   *RSFlipFlop
	rAnd *ANDGate
	sAnd *ANDGate
	Q    *NORGate
	QBar *NORGate
}

func NewLevelTriggeredDTypeLatch(clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatch {
	latch := &LevelTriggeredDTypeLatch{}

	latch.rAnd = NewANDGate(clkInPin, NewInverter(dataInPin))
	latch.sAnd = NewANDGate(clkInPin, dataInPin)

	latch.rs = NewRSFlipFLop(latch.rAnd, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

/*
func (l *LevelTriggeredDTypeLatch) UpdateDataPin(dataPin pwrEmitter) {
	l.rAnd.UpdatePin(2, 2, NewInverter(dataPin))
	l.sAnd.UpdatePin(2, 2, dataPin)
}
*/
type NBitLatch struct {
	latches []*LevelTriggeredDTypeLatch
	Qs      []pwrEmitter
}

func NewNBitLatch(clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLatch {
	latch := &NBitLatch{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatch(clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

// Level-triggered D-Type Latch With Clear ("Level" = clock high/low, "D" = data 0/1)

// d clk clr   q  !q
// 0  1   0    0  1
// 1  1   0    1  0
// X  0   0    q  !q  (data doesn't matter, no clock high to trigger a store-it action)
// X  X   1    0  1   (clear forces reset so data and clock do not matter)
// 1  1   1    X  X   (invalid)

type LevelTriggeredDTypeLatchWithClear struct {
	rs    *RSFlipFlop
	rAnd  *ANDGate
	sAnd  *ANDGate
	clrOR *ORGate
	Q     *NORGate
	QBar  *NORGate
}

func NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatchWithClear {
	latch := &LevelTriggeredDTypeLatchWithClear{}

	latch.rAnd = NewANDGate(clkInPin, NewInverter(dataInPin))
	latch.sAnd = NewANDGate(clkInPin, dataInPin)

	latch.clrOR = NewORGate(clrPin, latch.rAnd)

	latch.rs = NewRSFlipFLop(latch.clrOR, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

func (l *LevelTriggeredDTypeLatchWithClear) UpdateDataPin(dataPin pwrEmitter) {
	l.rAnd.UpdatePin(2, 2, NewInverter(dataPin))
	l.sAnd.UpdatePin(2, 2, dataPin)
}

type NBitLatchWithClear struct {
	latches []*LevelTriggeredDTypeLatchWithClear
	Qs      []pwrEmitter
}

func NewNBitLatchWithClear(clrPin, clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLatchWithClear {
	latch := &NBitLatchWithClear{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

func (l *NBitLatchWithClear) UpdateDataPins(dataPins []pwrEmitter) {

	// TODO: validate dataPins is same length as the latches slice

	for i, latch := range l.latches {
		latch.UpdateDataPin(dataPins[i])
	}
}

// Edge-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk   q  !q
// 0  ^    0  1
// 1  ^    1  0
// X  1    q  !q  (data doesn't matter, no clock transition to trigger a store-it action)
// X  0    q  !q  (data doesn't matter, no clock transition to trigger a store-it action)

type EdgeTriggeredDTypeLatch struct {
	lRAnd *ANDGate
	lSAnd *ANDGate
	lRS   *RSFlipFlop
	rRAnd *ANDGate
	rSAnd *ANDGate
	rRS   *RSFlipFlop
	Q     *NORGate
	QBar  *NORGate
}

func NewEdgeTriggeredDTypeLatch(clkInPin, dataInPin pwrEmitter) *EdgeTriggeredDTypeLatch {
	latch := &EdgeTriggeredDTypeLatch{}

	latch.rRAnd = NewANDGate(clkInPin, nil)
	latch.rSAnd = NewANDGate(clkInPin, nil)
	latch.rRS = NewRSFlipFLop(latch.rRAnd, latch.rSAnd)

	latch.lRAnd = NewANDGate(NewInverter(clkInPin), dataInPin)
	latch.lSAnd = NewANDGate(NewInverter(clkInPin), NewInverter(dataInPin))
	latch.lRS = NewRSFlipFLop(latch.lRAnd, latch.lSAnd)

	latch.rRAnd.UpdatePin(2, 2, latch.lRS.Q)
	latch.rSAnd.UpdatePin(2, 2, latch.lRS.QBar)

	// refer to the inner-right-flipflop's outputs for easier external access
	latch.Q = latch.rRS.Q
	latch.QBar = latch.rRS.QBar

	return latch
}

func (l *EdgeTriggeredDTypeLatch) UpdateDataPin(dataPin pwrEmitter) {
	l.lRAnd.UpdatePin(2, 2, dataPin)
	l.lSAnd.UpdatePin(2, 2, NewInverter(dataPin))
}

func (l *EdgeTriggeredDTypeLatch) StateDump() string {

	state := fmt.Sprintf("Left_R_AND:    %t\n", l.lRAnd.GetIsPowered())
	state += fmt.Sprintf("Left_S_AND:    %t\n", l.lSAnd.GetIsPowered())
	state += fmt.Sprintf("Left_RS_Q:     %t\n", l.lRS.Q.GetIsPowered())
	state += fmt.Sprintf("Left_RS_QBar:  %t\n", l.lRS.QBar.GetIsPowered())

	state += fmt.Sprintf("Right_R_AND:   %t\n", l.rRAnd.GetIsPowered())
	state += fmt.Sprintf("Right_S_AND:   %t\n", l.rSAnd.GetIsPowered())
	state += fmt.Sprintf("Right_RS_Q:    %t\n", l.rRS.Q.GetIsPowered())
	state += fmt.Sprintf("Right_RS_QBar: %t\n", l.rRS.QBar.GetIsPowered())

	return state
}

// Frequency Divider

type FrequencyDivider struct {
	latch *EdgeTriggeredDTypeLatch
	Q     *NORGate
}

func NewFrequencyDivider(oscillator pwrEmitter) *FrequencyDivider {
	freqDiv := &FrequencyDivider{}

	freqDiv.latch = NewEdgeTriggeredDTypeLatch(oscillator, nil)
	freqDiv.latch.UpdateDataPin(freqDiv.latch.QBar)

	// refer to the inner-right-flipflop's outputs for easier external access
	freqDiv.Q = freqDiv.latch.Q

	return freqDiv
}
