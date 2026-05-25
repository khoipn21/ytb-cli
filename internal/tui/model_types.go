package tui

type screen string

const (
	screenSetup    screen = "setup"
	screenDownload screen = "download"
)

type downloadPanel int

const (
	panelMenu downloadPanel = iota
	panelTable
	panelDetail
)
