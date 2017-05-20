package circuit

import (
	"errors"
	"fmt"
	"regexp"
)

type onesComplimenter struct {
	xorGates []emitter
}

func NewOnesComplimenter(bits []byte, signal emitter) (*onesComplimenter, error) {

	match, err := regexp.MatchString("^[01]+$", string(bits))
	if err != nil {
		return nil, err
	}

	if !match {
		err = errors.New(fmt.Sprint("Input bits not in binary format: " + string(bits)))
		return nil, err
	}

	c := &onesComplimenter{}

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

func (c *onesComplimenter) Compliment() string {
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
