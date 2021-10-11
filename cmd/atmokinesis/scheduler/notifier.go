package scheduler

type Notifier interface {
	Notify()
	Wait() // Blocking function
}

func NewDefaultNotifier() Notifier {
	return &defaultNotifier{notifyChannel: make(chan byte, 1)}
}

type defaultNotifier struct {
	notifyChannel chan byte
}

func (d *defaultNotifier) Notify() {
	d.notifyChannel <- 0
}

func (d *defaultNotifier) Wait() {
	<-d.notifyChannel
}
