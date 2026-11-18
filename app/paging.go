package app

import (
	"tedi/config"
	"tedi/utils"
)

type pager[T any] interface {
	GoPrev()
	GoNext()
	GetTotalPage() int
	GetCurruntPagePointer() int
	GetCurruntPage() []T
}

type pagedList[T any] struct {
	ui      config.UIConfiguration
	limit   int
	measure func(item T, screenW int) int

	pager[T]

	items   []T
	cursor  int
	screenW int
	screenH int
}

func newPagedList[T any](ui config.UIConfiguration, limit int, measure func(item T, screenW int) int) *pagedList[T] {
	l := &pagedList[T]{ui: ui, limit: limit, measure: measure}
	l.layout()
	return l
}

func (l *pagedList[T]) setItems(items []T) {
	l.items = items
	l.layout()
}

func (l *pagedList[T]) resize(screenW, screenH int) {
	l.screenW, l.screenH = screenW, screenH
	l.layout()
}

func (l *pagedList[T]) height(item T) int {
	if l.ui.Layout == config.LayoutFit {
		return min(l.measure(item, l.screenW), max(l.screenH-1, 1))
	}
	return l.ui.MaxItemHeight
}

func (l *pagedList[T]) layout() {
	pos := 0
	if l.pager != nil {
		pos = l.GetCurruntPagePointer()
	}
	canvas := max(l.screenH-1, 1)

	if l.ui.Layout == config.LayoutFit {
		p := utils.NewItemSizePagination[T](canvas, 0)
		for _, item := range l.items {
			p.PutItem(item, l.height(item))
		}
		l.pager = p
	} else {
		p := utils.NewItemAutoDemandPagination[T](l.limit, max(canvas/l.ui.MaxItemHeight, 1))
		for _, item := range l.items {
			p.PutItem(item)
		}
		l.pager = p
	}

	for i := 0; i < pos; i++ {
		l.pager.GoNext()
	}
	l.cursor = min(l.cursor, max(len(l.GetCurruntPage())-1, 0))
}

func (l *pagedList[T]) GoPrev() {
	l.pager.GoPrev()
	l.cursor = 0
}

func (l *pagedList[T]) GoNext() {
	l.pager.GoNext()
	l.cursor = 0
}

func (l *pagedList[T]) moveCursor(delta int) {
	cursor := l.cursor + delta
	page := l.GetCurruntPagePointer()

	switch {
	case cursor < 0:
		l.pager.GoPrev()
		if l.GetCurruntPagePointer() != page {
			cursor = len(l.GetCurruntPage()) - 1
		}
	case cursor >= len(l.GetCurruntPage()):
		l.pager.GoNext()
		if l.GetCurruntPagePointer() != page {
			cursor = 0
		}
	}

	l.cursor = min(max(cursor, 0), max(len(l.GetCurruntPage())-1, 0))
}

func (l *pagedList[T]) selected() (T, bool) {
	page := l.GetCurruntPage()
	if l.cursor >= len(page) {
		var zero T
		return zero, false
	}
	return page[l.cursor], true
}
