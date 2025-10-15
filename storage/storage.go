package storage

type Storage interface {
	Get(key string) (string, error)
	Set(key, value string) error
}
