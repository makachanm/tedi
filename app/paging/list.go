package paging

import "tedi/config"

type pager[T any] interface {
	GoPrev()
	GoNext()
	GetTotalPage() int
	GetCurruntPagePointer() int
	GetCurruntPage() []T
}

type PagedList[T any] struct {
	ui      config.UIConfiguration
	limit   int
	measure func(item T, screenW int) int

	pager[T]

	items   []T
	Cursor  int
	screenW int
	screenH int
}

func NewPagedList[T any](ui config.UIConfiguration, limit int, measure func(item T, screenW int) int) *PagedList[T] {
	l := &PagedList[T]{ui: ui, limit: limit, measure: measure}
	l.Layout()
	return l
}

func (l *PagedList[T]) SetItems(items []T) {
	l.items = items
	l.Layout()
}

func (l *PagedList[T]) Resize(screenW, screenH int) {
	l.screenW, l.screenH = screenW, screenH
	l.Layout()
}

func (l *PagedList[T]) Height(item T) int {
	if l.ui.Layout == config.LayoutFit {
		return min(l.measure(item, l.screenW), max(l.screenH-1, 1))
	}
	return l.ui.MaxItemHeight
}

func (l *PagedList[T]) Layout() {
	pos := 0
	if l.pager != nil {
		pos = l.GetCurruntPagePointer()
	}
	canvas := max(l.screenH-1, 1)

	if l.ui.Layout == config.LayoutFit {
		p := NewItemSizePagination[T](canvas, 0)
		for _, item := range l.items {
			p.PutItem(item, l.Height(item))
		}
		l.pager = p
	} else {
		p := NewItemAutoDemandPagination[T](l.limit, max(canvas/l.ui.MaxItemHeight, 1))
		for _, item := range l.items {
			p.PutItem(item)
		}
		l.pager = p
	}

	for i := 0; i < pos; i++ {
		l.pager.GoNext()
	}
	l.Cursor = min(l.Cursor, max(len(l.GetCurruntPage())-1, 0))
}

func (l *PagedList[T]) GoPrev() {
	l.pager.GoPrev()
	l.Cursor = 0
}

func (l *PagedList[T]) GoNext() {
	l.pager.GoNext()
	l.Cursor = 0
}

func (l *PagedList[T]) MoveCursor(delta int) {
	Cursor := l.Cursor + delta
	page := l.GetCurruntPagePointer()

	switch {
	case Cursor < 0:
		l.pager.GoPrev()
		if l.GetCurruntPagePointer() != page {
			Cursor = len(l.GetCurruntPage()) - 1
		}
	case Cursor >= len(l.GetCurruntPage()):
		l.pager.GoNext()
		if l.GetCurruntPagePointer() != page {
			Cursor = 0
		}
	}

	l.Cursor = min(max(Cursor, 0), max(len(l.GetCurruntPage())-1, 0))
}

func (l *PagedList[T]) Selected() (T, bool) {
	page := l.GetCurruntPage()
	if l.Cursor >= len(page) {
		var zero T
		return zero, false
	}
	return page[l.Cursor], true
}
