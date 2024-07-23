package pool

type Pool[T any] struct {
	Items   []*T
	NewItem func() (*T, error)
	Reset   func(*T)
}

func New[T any](
	cap int,
	newItem func() (*T, error),
	reset func(*T),
) Pool[T] {
	return Pool[T]{
		Items:   make([]*T, 0, cap),
		NewItem: newItem,
		Reset:   reset,
	}
}

func (p *Pool[T]) Get() (*T, error) {
	if len(p.Items) == 0 {
		return p.NewItem()
	}

	item := p.Items[len(p.Items)-1]
	p.Items = p.Items[:len(p.Items)-1]
	return item, nil
}

func (p *Pool[T]) Put(item *T) {
	p.Reset(item)
	p.Items = append(p.Items, item)
}
