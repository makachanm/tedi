package utils

type ItemAutoDemandPagination[T any] struct {
	max_item_limit    int
	max_item_per_page int
	items             []T

	currunt_page_pointer int
	total_page_count     int
}

func NewItemAutoDemandPagination[T any](max_item_limit int, max_per_page_limit int) *ItemAutoDemandPagination[T] {
	ptk := new(ItemAutoDemandPagination[T])
	ptk.max_item_limit = max_item_limit
	ptk.max_item_per_page = max(max_per_page_limit, 1)
	ptk.items = make([]T, 0)

	ptk.currunt_page_pointer = 0
	ptk.total_page_count = 1

	return ptk
}

func (p *ItemAutoDemandPagination[T]) PutItem(item T) {
	if len(p.items) >= p.max_item_limit {
		//shift
		tmp := p.items[1:]

		p.items = make([]T, p.max_item_limit)
		copy(p.items, tmp)
		p.items[p.max_item_limit-1] = item
	} else {
		p.items = append(p.items, item)
	}

	p.total_page_count = max((len(p.items)+p.max_item_per_page-1)/p.max_item_per_page, 1)
}

func (p *ItemAutoDemandPagination[T]) GoPrev() {
	if p.currunt_page_pointer <= 0 {
		return
	} else {
		p.currunt_page_pointer--
	}
}

func (p *ItemAutoDemandPagination[T]) GoNext() {
	if p.currunt_page_pointer+1 >= p.total_page_count {
		return
	} else {
		p.currunt_page_pointer++
	}
}

func (p *ItemAutoDemandPagination[T]) GetTotalPage() int {
	return p.total_page_count - 1
}

func (p *ItemAutoDemandPagination[T]) GetCurruntPagePointer() int {
	return p.currunt_page_pointer
}

func (p *ItemAutoDemandPagination[T]) GetCurruntPage() []T {
	start_pos := p.currunt_page_pointer * p.max_item_per_page
	if start_pos >= len(p.items) {
		return make([]T, 0)
	}
	end_pos := min(start_pos+p.max_item_per_page, len(p.items))
	return p.items[start_pos:end_pos]
}
