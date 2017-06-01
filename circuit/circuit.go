package circuit

type powerPublisher interface {
	Subscribe(func(bool)) // who do I tell
	Publish(bool)             // how do I tell them
}

type publication struct {
	subscriberCallbacks []func(on bool)
}

func (p *publication) Subscribe(callback func(bool)) {
	p.subscriberCallbacks = append(p.subscriberCallbacks, callback)
}

func (p *publication) Publish(state bool) {
	for _, callback := range p.subscriberCallbacks {
		callback(state)
	}
}
