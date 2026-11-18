package app

import "sync"

type historyEntry struct {
	scene   string
	restore func()
}

type History struct {
	lock    sync.Mutex
	entries []historyEntry
}

var history = new(History)

func (h *History) forward(ctx ApplicationContext, to string, restore func()) {
	h.lock.Lock()
	h.entries = append(h.entries, historyEntry{ctx.GetCurruntScene(), restore})
	h.lock.Unlock()

	ctx.TranslateTo(to)
}

func (h *History) back(ctx ApplicationContext) {
	h.lock.Lock()
	if len(h.entries) == 0 {
		h.lock.Unlock()
		ctx.TranslateTo("main")
		return
	}
	entry := h.entries[len(h.entries)-1]
	h.entries = h.entries[:len(h.entries)-1]
	h.lock.Unlock()

	if entry.restore != nil {
		entry.restore()
	}
	ctx.TranslateTo(entry.scene)
}
