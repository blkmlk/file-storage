package cache

import "errors"

var (
	ErrExists = errors.New("key exists")
)

type Cache interface {
	Lock(keys []string) error
	Unlock(keys []string)
}
