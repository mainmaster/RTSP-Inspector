package ui

import (
	"rtsp-inspector/clients/rtsp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func StartApp() {
	myApp := app.New()
	window := myApp.NewWindow("RTSP Inspector")

	ui := NewUIWidgets()
	client := &rtsp.Client{}
	h := &Handlers{UI: ui, client: client}

	// Привязываем события
	ui.BtnOpen.OnTapped = h.HandleConnect
	//ui.BtnPlay.OnTapped = h.HandlePlay

	// Сборка интерфейса по частям
	content := container.NewBorder(
		makeTopPanel(ui),
		makeBottomPanel(ui),
		nil,
		nil,
		makeCenterContent(ui),
	)

	// Оборачиваем всё в Padded, чтобы контент не лип к рамке окна
	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(1000, 600))
	window.ShowAndRun()
}

func makeTopPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewBorder(
		nil,         // top
		nil,         // bottom
		nil,         // left
		ui.BtnOpen,  // right (кнопка прижмется к правому краю)
		ui.URLEntry, // center (растянется на всю ширину)
	)
}

func makeCenterContent(ui *Widgets) fyne.CanvasObject {
	// 1. Верхняя часть (логи)
	top := ui.LogScroll

	// 2. Нижняя часть:
	// Убираем внутренний скролл вокруг формы.
	// Просто кладем форму в HBox, чтобы она могла расти вширь.
	gridHoriz := container.NewHBox(
		ui.StatsForm,
		//layout.NewSpacer(), // Это прижмет форму влево, но даст ей расти
	)

	// Прижимаем К ВЕРХУ
	gridFixed := container.NewVBox(gridHoriz, layout.NewSpacer())

	// 3. САМЫЙ ВАЖНЫЙ МОМЕНТ:
	// Используем ОБЫЧНЫЙ NewScroll (не VScroll).
	// Он позволяет скроллить и по вертикали, и по горизонтали.
	// Оборачиваем его в NewStack, чтобы он занял ВСЮ выделенную VSplit область.
	bottom := container.NewStack(container.NewScroll(gridFixed))

	// 4. Разделитель
	split := container.NewVSplit(top, bottom)
	split.Offset = 0.5
	return split
}

func makeBottomPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),

		ui.BtnDescribe,
		layout.NewSpacer(),
	)
}
