package features

type Feature interface {
	Enable() error
	Disable() error
}
