package app

import (
	"termfedi/config"
	"termfedi/utils"
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

	items   []T
	pages   pager[T]
	screenW int
	screenH int
}

func newPagedList[T any](ui config.UIConfiguration, limit int, measure func(item T, screenW int) int) *pagedList[T] {
	return &pagedList[T]{ui: ui, limit: limit, measure: measure}
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
	if l.pages != nil {
		pos = l.pages.GetCurruntPagePointer()
	}
	canvas := max(l.screenH-1, 1)

	if l.ui.Layout == config.LayoutFit {
		p := utils.NewItemSizePagination[T](canvas, 0)
		for _, item := range l.items {
			p.PutItem(item, l.height(item))
		}
		l.pages = p
	} else {
		p := utils.NewItemAutoDemandPagination[T](l.limit, max(canvas/l.ui.MaxItemHeight, 1))
		for _, item := range l.items {
			p.PutItem(item)
		}
		l.pages = p
	}

	for i := 0; i < pos; i++ {
		l.pages.GoNext()
	}
}

func (l *pagedList[T]) page() []T {
	if l.pages == nil {
		return nil
	}
	return l.pages.GetCurruntPage()
}

func (l *pagedList[T]) prev() {
	if l.pages != nil {
		l.pages.GoPrev()
	}
}

func (l *pagedList[T]) next() {
	if l.pages != nil {
		l.pages.GoNext()
	}
}

func (l *pagedList[T]) position() (int, int) {
	if l.pages == nil {
		return 1, 1
	}
	return l.pages.GetCurruntPagePointer() + 1, l.pages.GetTotalPage() + 1
}
