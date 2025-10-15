package dispatcher

type Middleware interface {
	Process(update interface{}) error
}
