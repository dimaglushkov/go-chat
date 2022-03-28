package tui

import (
	"github.com/dimaglushkov/go-chat/chat"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
	"strconv"
	"sync"
	"unicode"
)

type Application struct {
	App           *tview.Application
	pagesMu       sync.Mutex
	pages         *tview.Pages
	Addr          serverAddr
	butler        chat.ButlerClient
	grpcConnector func(addr, port string) (chat.ButlerClient, error)
	gr            grpcReq
	roomPort      chat.RoomPort
	tcpConnector  func(addr, port string) (chat.ButlerClient, error)
}

func NewApp(
	grpcConnector func(addr, port string) (chat.ButlerClient, error),
) *Application {
	app := Application{}
	//app.loaded = make(chan any)
	app.grpcConnector = grpcConnector
	app.App = tview.NewApplication()
	app.App.EnableMouse(true)
	app.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			app.App.Stop()
		}
		return event
	})

	app.pages = tview.NewPages()
	app.initLoadingPage()
	app.initLobbyPage()
	app.initAddrPage()

	app.App.SetRoot(app.pages, true)
	return &app
}

func (app *Application) Stop(msg string) {
	app.App.Stop()
	log.Print(msg)
	os.Exit(0)
}

func (app *Application) initAddrPage() {
	addrPage := tview.NewForm().
		AddInputField("IP Addr", "", 25, nil, func(text string) {
			app.Addr.ipAddr = text
		}).
		AddInputField("Port", "", 25, nil, func(text string) {
			app.Addr.port = text
		}).
		AddButton("Submit", func() {
			if !app.Addr.validate() {
				return
			}
			if app.grpcConnector == nil {
				app.Stop("no grpc was provided to app")
			}
			go app.load("lobbyPage", "addrPage", func() (err error) {
				app.butler, err = app.grpcConnector(app.Addr.ipAddr, app.Addr.port)
				return err
			})
		}).
		AddButton("Quit", func() {
			app.App.Stop()
		})
	addrPage.SetTitle("Server address").
		SetBorder(true)

	app.pages.AddPage("addrPage", center(40, 9, addrPage), true, true)
}

func (app *Application) initLoadingPage() {
	loadingPage := tview.NewModal().
		SetText("Loading")
	app.pages.AddPage("loadingPage", loadingPage, true, false)
}

func (app *Application) load(nextPageName, fallbackPageName string, f func() error) {
	app.App.QueueUpdateDraw(func() {
		app.pages.SwitchToPage("loadingPage")
	})
	err := f()
	if err != nil {
		app.App.QueueUpdateDraw(func() {
			app.pages.SwitchToPage(fallbackPageName)
		})
		return
	}
	app.App.QueueUpdateDraw(func() {
		app.pages.SwitchToPage(nextPageName)
	})
}

func (app *Application) initLobbyPage() {
	lobbyPage := tview.NewForm()
	lobbyPage.AddInputField("room name", "", 20, func(text string, r rune) bool {
		if len(text) > 10 {
			return false
		}
		return true
	}, func(text string) {
		app.gr.Name = text
	})
	lobbyPage.AddDropDown("action", []string{
		"join",
		"create",
	}, 0, func(opt string, optId int) {
		if opt == "create" {
			if id := lobbyPage.GetFormItemIndex("room size"); id > -1 {
				return
			}
			lobbyPage.AddInputField("room size", "", 20,
				func(textToCheck string, lastChar rune) bool {
					if unicode.IsDigit(lastChar) && len(textToCheck) < 3 {
						return true
					}
					return false
				},
				func(text string) {
					temp, _ := strconv.ParseInt(text, 10, 32)
					app.gr.Size = int32(temp)
				})
		} else {
			if id := lobbyPage.GetFormItemIndex("room size"); id > -1 {
				lobbyPage.RemoveFormItem(id)
			}
		}
	})

	lobbyPage.AddButton("Submit", func() {
		if !app.gr.validate() {
			return
		}
		if app.grpcConnector == nil {
			app.Stop("no grpc was provided to app")
		}

	})
	lobbyPage.AddButton("Quit", func() {
		app.App.Stop()
	})

	lobbyPage.SetTitle("Connect or create a room").
		SetBorder(true)

	app.pages.AddPage("lobbyPage", center(38, 11, lobbyPage), true, false)
}
