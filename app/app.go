package app

import (
	"bufio"
	"context"
	"github.com/dimaglushkov/go-chat/server"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"sync"
	"unicode"
)

type Application struct {
	tviewApp     *tview.Application
	pages        *tview.Pages
	pageBuilders map[string]func() tview.Primitive
	Addr         serverAddr

	butler        server.ButlerClient
	butlerCon     *grpc.ClientConn
	grpcConnector func(addr, port string) (*grpc.ClientConn, error)

	tcpClient    *net.TCPConn
	tcpConnector func(addr, port string) (*net.TCPConn, error)

	rns      server.RoomNameSize
	roomPort *server.RoomPort
	action   string
	username string

	msgSender   *bufio.Writer
	msgReceiver *bufio.Scanner

	msgChatCancel chan struct{}
	msgRecDone    chan struct{}
	msgLock       sync.Mutex
	msgTable      *tview.Table
	msgCnt        int
}

func New() *Application {
	app := Application{}
	app.grpcConnector = grpcConnector
	app.tcpConnector = tcpConnector
	app.msgRecDone = make(chan struct{})

	app.tviewApp = tview.NewApplication().
		EnableMouse(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlQ || event.Key() == tcell.KeyCtrlC {
				app.stop()
			}
			return event
		})

	app.pages = tview.NewPages()
	app.pageBuilders = map[string]func() tview.Primitive{
		"loadingPage": app.newLoadingPage,
		"addrPage":    app.newAddrPage,
		"lobbyPage":   app.newLobbyPage,
		"chatPage":    app.newChatPage,
	}
	for pageName, pageFunc := range app.pageBuilders {
		if pageName != "chatPage" {
			app.pages.AddPage(pageName, pageFunc(), true, false)
		}
	}
	app.pages.ShowPage("addrPage")

	app.tviewApp.SetRoot(app.pages, true)
	return &app
}

func (app *Application) Run() error {
	return app.tviewApp.Run()
}

func (app *Application) stop() {
	app.closeChatPage()
	if app.butlerCon != nil {
		app.butlerCon.Close()
	}
	if app.tviewApp != nil {
		app.tviewApp.Stop()
	}
}

func (app *Application) newAddrPage() tview.Primitive {
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
				app.stop()
			}
			go app.load("lobbyPage", "addrPage", func() (err error) {
				app.butlerCon, err = app.grpcConnector(app.Addr.ipAddr, app.Addr.port)
				app.butler = server.NewButlerClient(app.butlerCon)
				return err
			})
		}).
		AddButton("Quit", func() {
			app.tviewApp.Stop()
		})
	addrPage.SetTitle("Server address").
		SetBorder(true)

	return center(40, 9, addrPage)
}

func (app *Application) newLobbyPage() tview.Primitive {
	lobbyPage := tview.NewForm()
	lobbyPage.AddInputField("user name", "", 20, func(text string, r rune) bool {
		if len(text) > 20 {
			return false
		}
		return true
	}, func(text string) {
		app.username = text
	})

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
		if app.tcpConnector == nil || len(app.username) < 2 {
			app.stop()
		}

		var err error
		if app.action == "create" {
			app.roomPort, err = app.butler.CreateRoom(context.Background(), &app.rns)
			if err != nil || app.roomPort == nil {
				return
			}
		} else if app.action == "join" {
			app.roomPort, err = app.butler.FindRoom(context.Background(), &server.RoomName{Name: app.rns.Name})
			if err != nil || app.roomPort == nil {
				return
			}
		}
		go app.load("chatPage", "lobbyPage", func() error {
			app.tcpClient, err = app.tcpConnector(app.Addr.ipAddr, strconv.FormatInt(int64(app.roomPort.Port), 10))
			return err
		})

	})
	lobbyPage.AddButton("Quit", func() {
		app.tviewApp.Stop()
	})

	lobbyPage.SetTitle("Connect or create a room").
		SetBorder(true)

	return center(38, 13, lobbyPage)
}

func (app *Application) newLoadingPage() tview.Primitive {
	loadingPage := tview.NewModal().
		SetText("Loading")
	return loadingPage
}

func (app *Application) load(nextPageName, fallbackPageName string, f func() error) {
	app.tviewApp.QueueUpdateDraw(func() {
		app.pages.SwitchToPage("loadingPage")
	})
	err := f()
	if err != nil {
		app.tviewApp.QueueUpdateDraw(func() {
			app.pages.SwitchToPage(fallbackPageName)
		})
		return
	}
	if !app.pages.HasPage(nextPageName) {
		app.pages.AddPage(nextPageName, app.pageBuilders[nextPageName](), true, false)
	}
	app.tviewApp.QueueUpdateDraw(func() {
		app.pages.SwitchToPage(nextPageName)
	})
}

func (app *Application) printMsg(msgText string) {
	app.msgLock.Lock()
	app.tviewApp.QueueUpdateDraw(func() {
		app.msgTable.SetCell(app.msgCnt, 0, &tview.TableCell{Text: msgText})
	})
	app.msgCnt++
	app.msgLock.Unlock()
}

func (app *Application) newChatPage() tview.Primitive {
	app.msgSender = bufio.NewWriter(app.tcpClient)
	app.msgReceiver = bufio.NewScanner(app.tcpClient)

	leftSideBar := newPrimitive()
	rightSideBar := newPrimitive()
	msgTable := tview.NewTable()
	msgTable.
		SetTitle("Chat room: " + app.rns.Name).
		SetTitleColor(tcell.ColorGreenYellow).
		SetBorder(true)
	msgInputField := tview.NewInputField()
	msgInputField.SetDoneFunc(func(key tcell.Key) {
		text := msgInputField.GetText()
		if len(text) == 0 {
			return
		}
		err := sendMsg(app.msgSender, text)
		if err != nil {
			return
		}
		go app.printMsg("me: " + text)
		msgInputField.SetText("")
	})
	msgTable.SetFocusFunc(func() {
		msgInputField.Focus(nil)
	})

	leaveButton := tview.NewButton("Leave").
		SetSelectedFunc(func() {
			app.closeChatPage()
			go app.tviewApp.QueueUpdateDraw(func() {
				app.pages.SwitchToPage("lobbyPage")
			})
		})

	chatPage := tview.NewGrid().
		SetRows(1, 0, 3).
		SetColumns(0, -4, 0).
		SetBorders(false).
		AddItem(msgInputField, 2, 1, 1, 1, 0, 0, true).
		AddItem(leaveButton, 0, 1, 1, 1, 0, 0, false)

	chatPage.AddItem(leftSideBar, 0, 0, 0, 0, 0, 0, false).
		AddItem(msgTable, 1, 0, 1, 3, 0, 0, false).
		AddItem(rightSideBar, 0, 0, 0, 0, 0, 0, false)

	chatPage.AddItem(leftSideBar, 1, 0, 1, 1, 0, 100, false).
		AddItem(msgTable, 1, 1, 1, 1, 0, 100, false).
		AddItem(rightSideBar, 1, 2, 1, 1, 0, 100, false)

	app.msgTable = msgTable
	_ = sendMsg(app.msgSender, app.username)
	go receiveMsg(app.msgReceiver, app.printMsg, app.msgRecDone)

	return chatPage
}

func (app *Application) closeChatPage() {
	if app.tcpClient != nil {
		app.tcpClient.Close()
	}
	<-app.msgRecDone

	app.msgLock.Lock()
	app.msgCnt = 0
	app.msgLock.Unlock()

	app.pages.RemovePage("chatPage")
}
