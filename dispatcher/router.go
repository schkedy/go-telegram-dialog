package dispatcher

type Router interface {
	HandleUpdate(update interface{}) error
}
