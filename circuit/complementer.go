package circuit

import (
	"errors"
	"fmt"
	"regexp"
)

type onesComplementer struct {
	xorGates []emitter
}

func NewOnesComplementer(bits []byte, signal emitter) (*onesComplementer, error) {

	match, err := regexp.MatchString("^[01]+$", string(bits))
	if err != nil {
		return nil, err
	}

	if !match {
		err = errors.New(fmt.Sprint("Input bits not in binary format: " + string(bits)))
		return nil, err
	}

	c := &onesComplementer{}

	for _, b := range bits {
		switch b {
		case '0':
			c.xorGates = append(c.xorGates, newXORGate(signal, nil))
		case '1':
			c.xorGates = append(c.xorGates, newXORGate(signal, &Battery{}))
		}
	}

	return c, nil
}

func (c *onesComplementer) Complement() string {
	s := ""

	for _, x := range c.xorGates {
		if x.Emitting() {
			s += "1"
		} else {
			s += "0"
		}
	}

	return s
}
