package features

type Feature interface {
	Enable() error
	Disable(options DisableOptions) error
}

type DisableOptions struct {
	ProjectName string
}
