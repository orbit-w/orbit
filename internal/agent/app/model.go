package agent

type Stopper interface {
	Stop() error
}
