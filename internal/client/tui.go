package client

import (
	"errors"
	"log"
	"path"
	"regexp"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/handlers"
	"github.com/size12/gophkeeper/internal/storage"
	"golang.design/x/clipboard"
)

// TUI is a struct for terminal user interface.
type TUI struct {
	*tview.Application
	pages  *tview.Pages
	Client *handlers.Client
}

// NewTUI gets new terminal user interface for client.
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

// authPage switches to authentication page, where user can log in or register.
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

// recordInfoPage switches to page, where are all records shown. You can choose one.
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

		list.AddItem(record.ID, record.Type.String()+" | "+record.Metadata, '*', f)
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

// recordPage switches to record page, where you can see decrypted record data, copy this data, or delete record.
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
		AddText(record.Metadata+" | "+record.Type.String(), true, tview.AlignCenter, tcell.ColorGreen).
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

// createTextRecord creates new text record.
func (app *TUI) createTextRecord() {
	record := entity.Record{Type: entity.TypeText}
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

// createCredentialsRecord creates credentials (login and password) record.
func (app *TUI) createCredentialsRecord() {
	record := entity.Record{Type: entity.TypeLoginAndPassword}
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

// createCardRecord creates credit card record (card number, expiration date, cvc).
func (app *TUI) createCardRecord() {
	record := entity.Record{Type: entity.TypeCreditCard}
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
		result, err := regexp.Match(`^(0[1-9]|1[0-2])[|/]?([0-9]{4}|[0-9]{2})$`, []byte(creditCard.ExpirationDate))
		if err != nil {
			log.Fatalln("Failed parse regex for check card expiration date.")
		}

		if !result {
			app.recordsInfoPage("Incorrect expiration date.")
			return
		}

		result, err = regexp.Match(`^(?:4[0-9]{12}(?:[0-9]{3})?|[25][1-7][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\d{3})\d{11})$`, []byte(creditCard.CardNumber))
		if err != nil {
			log.Fatalln("Failed parse regex for check card expiration date.")
			return
		}

		if !result {
			app.recordsInfoPage("Incorrect card number.")
			return
		}

		result, err = regexp.Match(`\d{3}`, []byte(creditCard.CVCCode))
		if err != nil {
			log.Fatalln("Failed parse regex for check card expiration date.")
		}

		if !result {
			app.recordsInfoPage("Incorrect CVC code.")
			return
		}

		record.Data, _ = creditCard.Bytes()
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

	app.pages.AddPage("createCardRecord", frame, true, true)
	app.pages.SwitchToPage("createCardRecord")
}

// createFileRecord creates file record. You can choose any file to save.
func (app *TUI) createFileRecord() {
	record := entity.Record{Type: entity.TypeFile}
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

// createRecordPage creates page, where you can choose record type, then create it.
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
