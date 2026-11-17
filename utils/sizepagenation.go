package utils

type ItemSizePagination[T any] struct {
	items map[int][]T

	canvas_size          int
	currunt_size         int
	currunt_page_pointer int
	total_page_count     int
	padding              int
}

func NewItemSizePagination[T any](canvas_height int, padding int) *ItemSizePagination[T] {
	ptk := new(ItemSizePagination[T])
	ptk.items = make(map[int][]T, 0)

	ptk.currunt_page_pointer = 0
	ptk.canvas_size = canvas_height
	ptk.padding = padding
	ptk.total_page_count = 0

	return ptk
}

func (p *ItemSizePagination[T]) PutItem(item T, height int) {
	if len(p.items) == 0 {
		p.items[0] = []T{item}
		p.currunt_size = height
		return
	}

	if p.canvas_size-p.currunt_size < height+p.padding {
		p.total_page_count++
		p.currunt_size = 0
	}

	p.items[p.total_page_count] = append(p.items[p.total_page_count], item)
	p.currunt_size += height
}

func (p *ItemSizePagination[T]) GoPrev() {
	if p.currunt_page_pointer <= 0 {
		return
	} else {
		p.currunt_page_pointer--
	}
}

func (p *ItemSizePagination[T]) GoNext() {
	if p.currunt_page_pointer >= p.total_page_count {
		return
	} else {
		p.currunt_page_pointer++
	}
}

func (p *ItemSizePagination[T]) GetTotalPage() int {
	return p.total_page_count
}

func (p *ItemSizePagination[T]) GetCurruntPagePointer() int {
	return p.currunt_page_pointer
}

func (p *ItemSizePagination[T]) GetCurruntPage() []T {
	return p.items[p.currunt_page_pointer]
}
