package circuit

type Battery struct {
}

func (b *Battery) Register(callback func(bool)) {
	callback(true)
}

//func (b *Battery) GetState() bool {
//	return true
//}

/*
func (b *Battery) Charge() {
	b.isPowered = true
	b.Publish()
}

func (b *Battery) Discharge() {
	b.isPowered = false
	b.Publish()
}
*/
// OLD STUFF
// OLD STUFF
// OLD STUFF
// OLD STUFF
// OLD STUFF
// OLD STUFF
type emitter interface {
	Emitting() bool
}

func (b *Battery) Emitting() bool {
	return true
}
