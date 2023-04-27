package client

import (
	"errors"
	"log"
	"path"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/handlers"
	"github.com/size12/gophkeeper/internal/storage"
	"golang.design/x/clipboard"
)

type TUI struct {
	*tview.Application
	pages  *tview.Pages
	Client *handlers.Client
}

func NewTUI(client *handlers.Client) *TUI {

	app := tview.NewApplication()
	pages := tview.NewPages()

	err := clipboard.Init()
	if err != nil {
		log.Fatalln("Failed init clipboard:", err)
	}

	app.SetRoot(pages, true).EnableMouse(true)

	tui := &TUI{
		Application: app,
		Client:      client,
		pages:       pages,
	}

	tui.authPage("")

	return tui
}

func (app *TUI) authPage(message string) {
	credentials := entity.UserCredentials{}
	form := tview.NewForm()

	form.AddInputField("Login", "", 20, nil, func(login string) {
		credentials.Login = login
	})

	form.AddPasswordField("Password", "", 20, '*', func(password string) {
		credentials.Password = password
	})

	form.AddPasswordField("Master Key", "", 20, '*', func(masterKey string) {
		credentials.MasterKey = []byte(masterKey)
	})

	form.AddButton("Login", func() {
		err := app.Client.Login(credentials)

		if errors.Is(err, storage.ErrWrongCredentials) {
			app.authPage("Wrong credentials. Please try again.")
			return
		}

		if errors.Is(err, handlers.ErrFieldIsEmpty) {
			app.authPage("Some fields are empty.")
			return
		}

		if errors.Is(err, storage.ErrUnknown) || err != nil {
			app.authPage("Something is wrong. Please try again later.")
			return
		}

		app.recordsInfoPage("Logged successfully.")
	})

	form.AddButton("Register", func() {
		err := app.Client.Register(credentials)

		if errors.Is(err, storage.ErrLoginExists) {
			app.authPage("Such login exists. Please try again.")
			return
		}

		if errors.Is(err, handlers.ErrFieldIsEmpty) {
			app.authPage("Some fields are empty.")
			return
		}

		if errors.Is(err, storage.ErrUnknown) || err != nil {
			app.authPage("Something is wrong. Please try again later.")
			return
		}

		app.recordsInfoPage("Registered successfully.")
	})

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("TAB - switch between fields | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText(message, false, tview.AlignRight, tcell.ColorWhite)

	app.pages.AddPage("authentication", frame, true, true)
	app.pages.SwitchToPage("authentication")
}

func (app *TUI) recordsInfoPage(message string) {
	records, err := app.Client.GetRecordsInfo()

	if errors.Is(err, storage.ErrUserUnauthorized) {
		app.authPage("Session expired. Please login again.")
		return
	}

	if errors.Is(err, storage.ErrUnknown) || err != nil {
		message = "Something is wrong. Please try later."
		return
	}

	list := tview.NewList()

	for _, record := range records {

		f := func(record entity.Record) func() {
			return func() {
				app.recordPage(record.ID, "")
			}
		}(record)

		if record.Metadata == "" {
			record.Metadata = "no metadata"
		}

		list.AddItem(record.ID, record.Type+" | "+record.Metadata, '*', f)
	}

	listFrame := tview.NewFrame(list).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("Up/Down - switch between records | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("Ctrl+N - create new record       | Ctrl+U - refresh", false, tview.AlignLeft, tcell.ColorWhite).
		AddText(message, false, tview.AlignRight, tcell.ColorWhite)

	listFrame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			app.createRecordPage("")
		}
		if event.Key() == tcell.KeyCtrlU {
			app.recordsInfoPage("Refreshed.")
		}
		return event
	})

	app.pages.AddPage("records", listFrame, true, true)
	app.pages.SwitchToPage("records")
}

func (app *TUI) recordPage(recordID string, message string) {
	record, err := app.Client.GetRecord(recordID)

	if errors.Is(err, storage.ErrUserUnauthorized) {
		app.authPage("Session expired. Please login again.")
		return
	}

	if errors.Is(err, storage.ErrNotFound) {
		app.recordsInfoPage("Not found this record.")
		return
	}

	if err != nil {
		app.recordsInfoPage("Failed get record.")
		return
	}

	if record.Metadata == "" {
		record.Metadata = "no metadata"
	}

	frame := tview.NewFrame(tview.NewTextView().SetText(string(record.Data)).SetTextColor(tcell.ColorYellow).SetDisabled(true)).SetBorders(0, 0, 0, 1, 4, 4).
		AddText(record.Metadata+" | "+record.Type, true, tview.AlignCenter, tcell.ColorGreen).
		AddText("Ctrl+K - copy | Ctrl+U - delete | ESC - return to the menu", false, tview.AlignLeft, tcell.ColorWhite).
		AddText(message, false, tview.AlignRight, tcell.ColorWhite)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.recordsInfoPage("Returned to menu.")
		}

		if event.Key() == tcell.KeyCtrlK {
			app.recordPage(recordID, "Copied successfully.")
			clipboard.Write(clipboard.FmtText, record.Data)
		}
		if event.Key() == tcell.KeyCtrlU {
			err := app.Client.DeleteRecord(recordID)

			if errors.Is(err, storage.ErrUserUnauthorized) {
				app.authPage("Session expired. Please login again.")
				return event
			}

			if errors.Is(err, handlers.ErrWrongMasterKey) {
				app.authPage("Wrong master key. Please login again.")
				return event
			}

			if errors.Is(err, storage.ErrNotFound) {
				app.recordsInfoPage("Failed to delete. Not found record.")
				return event
			}

			if errors.Is(err, storage.ErrUnknown) || err != nil {
				app.recordPage(recordID, "Something is wrong. Please try later.")
				return event

			}

			app.recordsInfoPage("Deleted successfully.")
		}
		return event
	})

	app.pages.AddPage("record", frame, true, true)
	app.pages.SwitchToPage("record")
}

func (app *TUI) createTextRecord() {
	record := entity.Record{Type: "TEXT"}
	form := tview.NewForm()

	form.AddTextArea("Text", "", 30, 5, 0, func(text string) {
		record.Data = []byte(text)
	})

	form.AddInputField("Metadata", "", 20, nil, func(text string) {
		record.Metadata = text
	})

	form.AddButton("OK", func() {
		err := app.Client.CreateRecord(record)

		if errors.Is(err, storage.ErrUserUnauthorized) {
			app.authPage("Session expired. Please login again.")
			return
		}

		if errors.Is(err, storage.ErrUnknown) {
			app.recordsInfoPage("Something is wrong. Please try later.")
			return
		}

		if errors.Is(err, handlers.ErrWrongMasterKey) {
			app.authPage("Wrong master key. Please login again.")
			return
		}

		app.recordsInfoPage("Created record successfully.")
	})

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("TAB - switch between fields | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("ESC - return to the menu.", false, tview.AlignLeft, tcell.ColorWhite)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.recordsInfoPage("Returned to menu.")
		}
		return event
	})

	app.pages.AddPage("createTextRecord", frame, true, true)
	app.pages.SwitchToPage("createTextRecord")
}

func (app *TUI) createCredentialsRecord() {
	record := entity.Record{Type: "LOGIN_AND_PASSWORD"}
	form := tview.NewForm()

	loginAndPassword := entity.LoginAndPassword{}

	form.AddInputField("Login", "", 20, nil, func(text string) {
		loginAndPassword.Login = text
	})

	form.AddInputField("Password", "", 20, nil, func(text string) {
		loginAndPassword.Password = text
	})

	form.AddInputField("Metadata", "", 20, nil, func(text string) {
		record.Metadata = text
	})

	form.AddButton("OK", func() {
		record.Data, _ = loginAndPassword.Bytes()
		err := app.Client.CreateRecord(record)

		if errors.Is(err, storage.ErrUserUnauthorized) {
			app.authPage("Session expired. Please login again.")
			return
		}

		if errors.Is(err, storage.ErrUnknown) {
			app.recordsInfoPage("Something is wrong. Please try later.")
			return
		}

		if errors.Is(err, handlers.ErrWrongMasterKey) {
			app.authPage("Wrong master key. Please login again.")
			return
		}

		app.recordsInfoPage("Created record successfully.")
	})

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("TAB - switch between fields | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("ESC - exit to all records.", false, tview.AlignLeft, tcell.ColorWhite)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.recordsInfoPage("Returned to menu.")
		}
		return event
	})

	app.pages.AddPage("createTextRecord", frame, true, true)
	app.pages.SwitchToPage("createTextRecord")
}

func (app *TUI) createCardRecord() {
	record := entity.Record{Type: "CREDIT_CARD"}
	form := tview.NewForm()

	creditCard := entity.CreditCard{}

	form.AddInputField("Number", "", 20, nil, func(text string) {
		creditCard.CardNumber = text
	})

	form.AddInputField("Expiration", "", 20, nil, func(text string) {
		creditCard.ExpirationDate = text
	})

	form.AddInputField("CVC", "", 3, nil, func(text string) {
		creditCard.CVCCode = text
	})

	form.AddInputField("Metadata", "", 20, nil, func(text string) {
		record.Metadata = text
	})

	form.AddButton("OK", func() {
		record.Data, _ = creditCard.Bytes()
		err := app.Client.CreateRecord(record)

		if errors.Is(err, storage.ErrUserUnauthorized) {
			app.authPage("Session expired. Please login again.")
			return
		}

		if errors.Is(err, storage.ErrUnknown) {
			app.recordsInfoPage("Something is wrong. Please try later.")
			return
		}

		if errors.Is(err, handlers.ErrWrongMasterKey) {
			app.authPage("Wrong master key. Please login again.")
			return
		}

		app.recordsInfoPage("Created record successfully.")
	})

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("TAB - switch between fields | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("ESC - return to the menu.", false, tview.AlignLeft, tcell.ColorWhite)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.recordsInfoPage("Returned to menu.")
		}
		return event
	})

	app.pages.AddPage("createCardRecord", frame, true, true)
	app.pages.SwitchToPage("createCardRecord")
}

func (app *TUI) createFileRecord() {
	record := entity.Record{Type: "FILE"}
	form := tview.NewForm()

	file := entity.BinaryFile{}

	form.AddInputField("Filepath", "", 20, nil, func(text string) {
		file.FilePath = text
	})

	form.AddButton("OK", func() {
		filename := path.Base(file.FilePath)
		record.Metadata = filename

		data, err := file.Bytes()

		if err != nil {
			app.recordsInfoPage("Failed opened file.")
			return
		}

		record.Data = data

		err = app.Client.CreateRecord(record)

		if errors.Is(err, storage.ErrUserUnauthorized) {
			app.authPage("Session expired. Please login again.")
			return
		}

		if errors.Is(err, storage.ErrUnknown) {
			app.recordsInfoPage("Something is wrong. Please try later.")
			return
		}

		if errors.Is(err, handlers.ErrWrongMasterKey) {
			app.authPage("Wrong master key. Please login again.")
			return
		}

		app.recordsInfoPage("Created record successfully.")
	})

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("TAB - switch between fields | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("ESC - return to the menu.", false, tview.AlignLeft, tcell.ColorWhite)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.recordsInfoPage("Returned to menu.")
		}
		return event
	})

	app.pages.AddPage("createFileRecord", frame, true, true)
	app.pages.SwitchToPage("createFileRecord")
}

func (app *TUI) createRecordPage(message string) {
	form := tview.NewForm()

	form.AddDropDown("Type", []string{"Text", "Login + password", "Credit card", "Binary file"}, -1, func(option string, optionIndex int) {
		switch option {
		case "Text":
			app.createTextRecord()
		case "Login + password":
			app.createCredentialsRecord()
		case "Credit card":
			app.createCardRecord()
		case "Binary file":
			app.createFileRecord()
		}
	})

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 1, 4, 4).
		AddText("TAB - switch between fields | Enter - choose this option", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("ESC - exit to all records.", false, tview.AlignLeft, tcell.ColorWhite).
		AddText(message, false, tview.AlignRight, tcell.ColorWhite)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.recordsInfoPage("Returned to menu.")
		}
		return event
	})

	app.pages.AddPage("create", frame, true, true)
	app.pages.SwitchToPage("create")
}
