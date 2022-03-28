package app

import (
	"bufio"
	"context"
	"github.com/dimaglushkov/go-chat/chat"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"unicode"
)

type Application struct {
	App          *tview.Application
	pages        *tview.Pages
	pageBuilders map[string]func() tview.Primitive
	Addr         serverAddr

	butler        chat.ButlerClient
	butlerCon     *grpc.ClientConn
	grpcConnector func(addr, port string) (chat.ButlerClient, *grpc.ClientConn, error)

	tcpClient    *net.TCPConn
	tcpConnector func(addr, port string) (*net.TCPConn, error)
	rns          chat.RoomNameSize

	roomPort *chat.RoomPort
	action   string
	username string

	msgSender   *bufio.Writer
	msgReceiver *bufio.Scanner
	msgCnt      int
	msgCntLock  sync.Mutex
}

func NewApp(
	grpcConnector func(addr, port string) (chat.ButlerClient, *grpc.ClientConn, error),
	tcpConnector func(addr, port string) (*net.TCPConn, error),
) *Application {
	app := Application{}

	app.grpcConnector = grpcConnector
	app.tcpConnector = tcpConnector

	app.App = tview.NewApplication()
	app.App.EnableMouse(true)
	app.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ || event.Key() == tcell.KeyCtrlC {
			app.Stop()
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

	app.App.SetRoot(app.pages, true)
	return &app
}

func (app *Application) Stop() {
	if app.tcpClient != nil {
		app.tcpClient.Close()
	}
	if app.butlerCon != nil {
		app.butlerCon.Close()
	}
	app.App.Stop()
	os.Exit(0)
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
				app.Stop()
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
			app.Stop()
		}

		var err error
		if app.action == "create" {
			app.roomPort, err = app.butler.CreateRoom(context.Background(), &app.rns)
			if err != nil || app.roomPort == nil {
				return
			}
		} else if app.action == "join" {
			app.roomPort, err = app.butler.FindRoom(context.Background(), &chat.RoomName{Name: app.rns.Name})
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
		app.App.Stop()
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
	if !app.pages.HasPage(nextPageName) {
		app.pages.AddPage(nextPageName, app.pageBuilders[nextPageName](), true, false)
	}
	app.App.QueueUpdateDraw(func() {
		app.pages.SwitchToPage(nextPageName)
	})
}

func (app *Application) sendMsg(msg string) {
	if app.msgSender == nil {
		log.Println("app.msgSender is nil")
		return
	}
	if len(msg) == 0 {
		log.Println("msg is empty")
		return
	}
	if msg[len(msg)-1] != '\n' {
		msg += "\n"
	}
	_, err := app.msgSender.WriteString(msg)
	if err != nil {
		log.Println(err)
	}
	err = app.msgSender.Flush()
	if err != nil {
		log.Println(err)
	}
}

func (app *Application) receiveMsg(msgTable *tview.Table) {
	if app.msgReceiver == nil {
		log.Println("app.msgReceiver is nil")
		return
	}
	for app.msgReceiver.Scan() {
		msgText := app.msgReceiver.Text()

		app.msgCntLock.Lock()
		msgTable.SetCell(app.msgCnt, 0, &tview.TableCell{Text: msgText})
		app.msgCnt++
		app.msgCntLock.Unlock()

		app.App.Draw()
	}

}

func (app *Application) newChatPage() tview.Primitive {
	app.msgSender = bufio.NewWriter(app.tcpClient)
	app.msgReceiver = bufio.NewScanner(app.tcpClient)

	newPrimitive := func() tview.Primitive {
		return tview.NewFrame(nil).
			SetBorders(0, 0, 0, 0, 0, 0)
	}

	leftSideBar := newPrimitive()
	rightSideBar := newPrimitive()
	msgTable := tview.NewTable()
	msgTable.
		SetTitle("Chat room: " + app.rns.Name).
		SetTitleColor(tcell.ColorGreenYellow).
		SetBorder(false)
	msgInputField := tview.NewInputField()
	msgInputField.SetDoneFunc(func(key tcell.Key) {
		msgText := msgInputField.GetText()
		app.sendMsg(msgText)

		app.msgCntLock.Lock()
		msgTable.SetCell(app.msgCnt, 0, &tview.TableCell{Text: "me: " + msgText})
		app.msgCnt++
		app.msgCntLock.Unlock()

		msgInputField.SetText("")
	})

	chatPage := tview.NewGrid().
		SetRows(1, 0, 3).
		SetColumns(0, -4, 0).
		SetBorders(true).
		AddItem(msgInputField, 2, 1, 1, 1, 0, 0, true)

	chatPage.AddItem(leftSideBar, 0, 0, 0, 0, 0, 0, false).
		AddItem(msgTable, 1, 0, 1, 3, 0, 0, false).
		AddItem(rightSideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	chatPage.AddItem(leftSideBar, 1, 0, 1, 1, 0, 100, false).
		AddItem(msgTable, 1, 1, 1, 1, 0, 100, false).
		AddItem(rightSideBar, 1, 2, 1, 1, 0, 100, false)

	app.sendMsg(app.username)
	go app.receiveMsg(msgTable)

	return chatPage
}
