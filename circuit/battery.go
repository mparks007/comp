package circuit

import (
	"sync/atomic"
)

// Battery is a low-tech power source to simply store/transmit power state of on or off
type Battery struct {
	pwrSource // battery gains all that is pwrSource too
}

// NewBattery will return a battery whose initial state is based on the passed in initialization value
func NewBattery(name string, startState bool) *Battery {
	bat := &Battery{}
	bat.Init()
	bat.Name = name

	bat.isPowered.Store(startState)
	return bat
}

// Charge will simulate a live battery by simply transmitting power as on (and tracking a unique sequence number for all downstream results of THIS state change)
func (b *Battery) Charge() {
	b.Transmit(true, atomic.AddInt64(&globalSeqNum, 1))
}

// Discharge will simulate a dead battery by simply transmitting power as off (and tracking a unique sequence number for all downstream results of THIS state change)
func (b *Battery) Discharge() {
	b.Transmit(false, atomic.AddInt64(&globalSeqNum, 1))
}
