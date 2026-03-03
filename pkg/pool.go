package pkg

import "sync"

// Resetter определяет интерфейс для типов, которые могут сбрасывать своё состояние
type Resetter interface {
	Reset()
}

// Pool представляет собой пул объектов с generic-параметром.
// Тип T должен реализовывать метод Reset().
// Pool использует sync.Pool для хранения и переиспользования объектов.
type Pool[T Resetter] struct {
	pool *sync.Pool
}

// New создаёт и возвращает новый пул для типа T.
// Функция создаёт sync.Pool с фабрикой, которая возвращает новый объект типа T.
func New[T Resetter](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

// Get возвращает объект из пула.
// Если пул пуст, создаётся новый объект с помощью функции New из sync.Pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put помещает объект в пул.
// Перед помещением объекта в пул вызывается метод Reset() для сброса состояния объекта.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
