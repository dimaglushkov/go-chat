package app

import (
	"context"
	"fmt"
	"github.com/dimaglushkov/go-chat/chat"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
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
	butlerCon     *grpc.ClientConn
	grpcConnector func(addr, port string) (chat.ButlerClient, *grpc.ClientConn, error)
	cc            ChatConn
	tcpConnector  func(addr, port string) (ChatConn, error)
	rns           chat.RoomNameSize
	action        string
	roomPort      *chat.RoomPort
}

func NewApp(
	grpcConnector func(addr, port string) (chat.ButlerClient, *grpc.ClientConn, error),
	tcpConnector func(addr, port string) (ChatConn, error),
) *Application {
	app := Application{}

	app.grpcConnector = grpcConnector
	app.tcpConnector = tcpConnector

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
	if app.cc != nil {
		app.butlerCon.Close()
	}
	if app.cc != nil {
		app.cc.Close()
	}
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
				app.butler, app.butlerCon, err = app.grpcConnector(app.Addr.ipAddr, app.Addr.port)
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

func (app *Application) initLobbyPage() {
	lobbyPage := tview.NewForm()
	lobbyPage.AddInputField("room name", "", 20, func(text string, r rune) bool {
		if len(text) > 10 {
			return false
		}
		return true
	}, func(text string) {
		app.rns.Name = text
	})
	lobbyPage.AddDropDown("action", []string{
		"join",
		"create",
	}, 0, func(opt string, optId int) {
		app.action = opt
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
					app.rns.Size = int32(temp)
				})
		} else {
			if id := lobbyPage.GetFormItemIndex("room size"); id > -1 {
				lobbyPage.RemoveFormItem(id)
			}
		}
	})

	lobbyPage.AddButton("Submit", func() {
		if app.tcpConnector == nil {
			app.Stop("no tcp was provided to app")
		}

		var err error
		if app.action == "create" {
			app.roomPort, err = app.butler.CreateRoom(context.Background(), &app.rns)
			if err != nil || app.roomPort == nil {
				log.Print("Room was not created: " + err.Error())
				return
			}
		} else if app.action == "join" {
			app.roomPort, err = app.butler.FindRoom(context.Background(), &chat.RoomName{Name: app.rns.Name})
			if err != nil || app.roomPort == nil {
				log.Print("Room was not found")
				return
			}
		}
		log.Print("Room found / created successfully")
		go app.load("chatPage", "lobbyPage", func() error {
			app.cc, err = app.tcpConnector(app.Addr.ipAddr, fmt.Sprint(app.roomPort.Port))
			return err
		})

	})
	lobbyPage.AddButton("Quit", func() {
		app.App.Stop()
	})

	lobbyPage.SetTitle("Connect or create a room").
		SetBorder(true)

	app.pages.AddPage("lobbyPage", center(38, 11, lobbyPage), true, false)
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
