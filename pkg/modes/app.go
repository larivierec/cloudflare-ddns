package modes

import "fmt"

type AppMode interface {
	Init() error
	Start() error
	Stop() error
}

func NewAppMode(mode string) (AppMode, error) {
	switch mode {
	case "application":
		return &Container{}, nil
	case "api":
		return &API{}, nil
	case "serverless":
		return &Serverless{}, nil
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}
}
