package circuit

type publisher interface {
	Publish()
	Register(func(bool))
}

type publication struct {
	state               bool
	subscriberCallbacks []func(on bool)
}

func (p *publication) Register(callback func(bool)) {
	p.subscriberCallbacks = append(p.subscriberCallbacks, callback)

	// ensure newly registered callback immediately gets current state
	callback(p.state)
}

func (p *publication) Publish() {
	for _, subscriber := range p.subscriberCallbacks {
		subscriber(p.state)
	}
}
