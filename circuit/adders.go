package circuit

// a b		sum		carry
// 0 0		0		0
// 1 0		1		0
// 0 1		1		0
// 1 1		0		1

/*********************************/
/********** HalfAdder ************/
/*********************************/

type halfAdder struct {
	sum   emitter
	carry emitter
}

func newHalfAdder(pin1 emitter, pin2 emitter) *halfAdder {
	return &halfAdder{
		newXORGate(pin1, pin2),
		newANDGate(pin1, pin2),
	}
}

// DELETE THESE TWO
func (h *halfAdder) Sum() bool {
	return h.sum.Emitting()
}

func (h *halfAdder) Carry() bool {
	return h.carry.Emitting()
}

/*********************************/
/********** FullAdder ************/
/*********************************/

type fullAdder struct {
	halfAdder1 *halfAdder
	halfAdder2 *halfAdder
	sum        emitter
	carry      emitter
}

func newFullAdder(pin1 emitter, pin2 emitter, carry emitter) *fullAdder {
	f := &fullAdder{}

	f.halfAdder1 = newHalfAdder(pin1, pin2)
	f.halfAdder2 = newHalfAdder(f.halfAdder1.sum, carry)
	f.sum = newContact(f.halfAdder2.sum)
	f.carry = newContact(f.halfAdder1.carry, f.halfAdder2.carry)

	return f
}

// DELETE THIS
func (f *fullAdder) Sum() bool {
	return f.halfAdder2.sum.Emitting()
}

func (f *fullAdder) Carry() bool {
	return f.halfAdder1.carry.Emitting() || f.halfAdder2.carry.Emitting()
}

/*********************************/
/*********** bitsAdder ***********/
/*********************************/

//    0110011101
// +  1011010110
// = 10001110011

type bitAdder struct {
	fullAdders []fullAdder
}

func NewBitAdder(b1, b2 string) *bitAdder {
	b := &bitAdder{}

	return b
}

func (b *bitAdder) String() string {
	return "answer"
}
