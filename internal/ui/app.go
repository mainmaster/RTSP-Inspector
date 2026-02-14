package ui

import (
	"fmt"
	"rtsp-inspector/clients/rtsp" // замените на ваш путь

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func StartApp() {
	myApp := app.New()
	window := myApp.NewWindow("RTSP Inspector")

	ui := NewUIWidgets()
	client := rtsp.Client{}
	// В идеале данные для SetCredentials тоже брать из UI
	client.SetCredentials(rtsp.Credentials{Username: "admin", Password: "password"})

	// Назначаем действия кнопкам
	ui.BtnOptions.OnTapped = func() {
		req, _ := rtsp.NewRequest("OPTIONS", ui.URLEntry.Text)
		res, err := client.Do(req)
		if err != nil {
			ui.AppendLog("Err: " + err.Error())
			return
		}
		ui.AppendLog(fmt.Sprintf("[OPTIONS] %v", res))
	}

	ui.BtnDesc.OnTapped = func() {
		req, _ := rtsp.NewRequest("DESCRIBE", ui.URLEntry.Text)
		res, err := client.Do(req)
		if err != nil {
			ui.AppendLog("Err: " + err.Error())
			return
		}
		ui.AppendLog(fmt.Sprintf("[DESCRIBE] SDP:\n%s", string(res.Body)))
	}

	ui.BtnClear.OnTapped = func() {
		ui.LogOutput.SetText("")
	}

	// Верстка интерфейса
	controls := container.NewVBox(
		ui.URLEntry,
		container.NewGridWithColumns(3, ui.BtnOptions, ui.BtnDesc, ui.BtnSetup),
	)

	footer := container.NewHBox(ui.BtnClear)

	// Основной контент: Сверху кнопки, в центре лог (со скроллом)
	content := container.NewBorder(
		controls,
		footer,
		nil, nil,
		container.NewScroll(ui.LogOutput),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}
