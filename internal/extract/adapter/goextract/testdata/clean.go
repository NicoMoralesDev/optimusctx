package fixtures

const Version = "v1"

var Enabled = true

type Widget struct {
	Name string
	Size int
}

type Handler interface {
	Handle(input string) error
}

func (w *Widget) Run(input string) error {
	return nil
}

func BuildWidget(name string) *Widget {
	return &Widget{Name: name}
}
