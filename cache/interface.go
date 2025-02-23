package cache

import (
	"time"
)

const (
	EXPIRE_KEY = "%s:expire_at"
)

type ICache interface {
	Set(string, interface{}, time.Duration) error
	Get(string) (interface{}, error)
	GetString(string) (string, error)
	Delete(string) error
}
