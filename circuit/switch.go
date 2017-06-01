package circuit

type Switch struct {
	switchOn bool
	//	switchIsPowered bool
	publication
}

/*
func NewSwitch(init bool, powerSource powerPublisher) *Switch {
	s := &Switch{}

	s.switchOn = init
	s.SubScribeToPower(powerSource)

	return s
}
*/
func NewSwitch(init bool) *Switch {
	s := &Switch{}

	s.switchOn = init

	return s
}

/*
func (s *Switch) SubScribeToPower(powerSource powerPublisher) {
	powerSource.Subscribe(s.powerUpdate)
}
*/

/*
func (s *Switch) powerUpdate(newState bool) {
	if s.switchIsPowered != newState {
		s.switchIsPowered = newState
		s.Publish(s.switchOn && newState)
	}
}
*/
func (s *Switch) TurnOn() {
	if !s.switchOn {
		s.switchOn = true
		s.Publish(true)
		//s.Publish(s.switchIsPowered)
	}
}

func (s *Switch) TurnOff() {
	if s.switchOn {
		s.switchOn = false
		s.Publish(false)
	}
}

func (s *Switch) Toggle() {
	if s.switchOn {
		s.TurnOff()
	} else {
		s.TurnOn()
	}
}
