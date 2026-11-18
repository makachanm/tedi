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

func (h *History) push(entry historyEntry) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.entries = append(h.entries, entry)
}

func (h *History) pop() (historyEntry, bool) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if len(h.entries) == 0 {
		return historyEntry{}, false
	}
	entry := h.entries[len(h.entries)-1]
	h.entries = h.entries[:len(h.entries)-1]
	return entry, true
}

func (m *MainAppContexts) Forward(scene string, restore func()) {
	m.history.push(historyEntry{*m.scene, restore})
	m.transision_signal <- scene
}

func (m *MainAppContexts) Back() {
	entry, ok := m.history.pop()
	if !ok {
		m.Exit(0)
		return
	}
	if entry.restore != nil {
		entry.restore()
	}
	m.transision_signal <- entry.scene
}
