package orivil

type Session interface {

	ID() string

	SetData(key string, data interface{})

	GetData(key string) interface{}

	FlashData(key string) (data interface{})

	DelData(key string)

	Set(key, value string)

	Get(key string) (value string)

	Flash(key string) (value string)

	Del(key string)
}

type PSession interface {

	ID() string

	Set(key, value string)

	Get(key string) (value string)

	Flash(key string) (value string)

	Del(key string)
}
