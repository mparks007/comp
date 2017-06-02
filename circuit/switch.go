package circuit

type Switch struct {
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
	//s := &Switch{{init, nil}}  // why cant I initialize like this?
	s := &Switch{}

	s.state = init

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
	if !s.state {
		s.state = true
		s.Publish()
		//s.Publish(s.switchIsPowered)
	}
}

func (s *Switch) TurnOff() {
	if s.state {
		s.state = false
		s.Publish()
	}
}

func (s *Switch) Toggle() {
	if s.state {
		s.TurnOff()
	} else {
		s.TurnOn()
	}
}
