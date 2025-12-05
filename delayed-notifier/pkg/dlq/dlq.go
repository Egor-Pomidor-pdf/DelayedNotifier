package dlq

type DLQ[T any] struct {
	items chan *Item[T]
}

type Item[T any] struct {
	value T
	err error
}

func (i * Item[T]) Value() T {
	return i.value
}

func (i *Item[T]) Error() error {
	return i.err
}

func NewDLQ[T any](bufferSize int) *DLQ[T] {
	return &DLQ[T]{
		items: make(chan *Item[T], bufferSize),
	}
}

func (d * DLQ[T]) Put(value T, err error) {
	d.items <- &Item[T]{value: value, err: err}
} 

func (d *DLQ[T]) Items() <-chan *Item[T] {
	return d.items
}

func (d *DLQ[T]) Close() {
	close(d.items)
}