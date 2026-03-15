package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func CreateAndRun(window fyne.Window, widgets *Widgets) {
	content := container.NewBorder(
		makeTopPanel(widgets),
		makeBottomPanel(),
		nil,
		nil,
		makeCenterContent(widgets),
	)

	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(1000, 600))
	window.ShowAndRun()
}

func makeTopPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		nil,
		ui.BtnConnect,
		ui.urlEntry,
	)
}

func makeCenterContent(ui *Widgets) fyne.CanvasObject {
	logSplit := container.NewHSplit(
		ui.logList,
		ui.detailView,
	)
	logSplit.Offset = 0.3

	leftStats := container.NewVScroll(ui.rtpForm)
	rightStats := container.NewVScroll(ui.naluForm)
	bottomSplit := container.NewHSplit(leftStats, rightStats)
	bottomSplit.Offset = 0.3

	mainSplit := container.NewVSplit(logSplit, bottomSplit)
	mainSplit.Offset = 0.7

	return mainSplit
}

func makeBottomPanel() fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),
		layout.NewSpacer(),
	)
}
