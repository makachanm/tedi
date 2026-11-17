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
		l.GoNext()
	}
}
