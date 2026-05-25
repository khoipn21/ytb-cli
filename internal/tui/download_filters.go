package tui

func (m Model) selectedVideo() (videoState, bool) {
	idx, ok := m.selectedVideoIndex()
	if !ok {
		return videoState{}, false
	}
	return m.videos[idx], true
}

func (m Model) selectedVideoIndex() (int, bool) {
	if len(m.visibleRows) == 0 {
		return -1, false
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.visibleRows) {
		return -1, false
	}
	idx := m.visibleRows[cursor]
	if idx < 0 || idx >= len(m.videos) {
		return -1, false
	}
	return idx, true
}

func (m Model) detailVideoWithIndex() (videoState, int, bool) {
	if m.detailIndex < 0 || m.detailIndex >= len(m.videos) {
		return videoState{}, -1, false
	}
	return m.videos[m.detailIndex], m.detailIndex, true
}

func (m Model) matchesFilter(status string) bool {
	switch m.filterIndex {
	case 1:
		return status == "downloading"
	case 2:
		return status == "done"
	case 3:
		return status == "error"
	case 4:
		return status == "queued"
	default:
		return true
	}
}

func (m Model) resolveVideoStatus(index int, item videoState) string {
	if item.hasError {
		return "error"
	}
	if item.done {
		return "done"
	}
	if index == m.currentIndex {
		return "downloading"
	}
	return "queued"
}
