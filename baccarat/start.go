package baccarat

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/SixofClubsss/Holdero/holdero"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
)

const app_tag = "Baccarat"

// Run Baccarat as a single dApp
func StartApp() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	menu.InitLogrusLog(runtime.GOOS == "windows")
	config := menu.ReadDreamsConfig(app_tag)

	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.SetIcon(holdero.ResourcePokerBotIconPng)
	w.Resize(fyne.NewSize(1400, 800))
	w.SetMaster()
	done := make(chan struct{})

	dreams.Theme.Img = *canvas.NewImageFromResource(nil)
	d := dreams.AppObject{
		Window:     w,
		Background: container.NewMax(&dreams.Theme.Img),
	}
	d.SetChannels(1)

	closeFunc := func() {
		menu.WriteDreamsConfig(
			dreams.SaveData{
				Skin:   config.Skin,
				Daemon: []string{rpc.Daemon.Rpc},
				DBtype: menu.Gnomes.DBType,
			})
		menu.Gnomes.Stop(app_tag)
		d.StopProcess()
		w.Close()
	}

	w.SetCloseIntercept(func() { closeFunc() })

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		closeFunc()
	}()

	rpc.InitBalances()

	go func() {
		time.Sleep(3 * time.Second)
		ticker := time.NewTicker(6 * time.Second)
		for {
			select {
			case <-ticker.C:
				rpc.Ping()
				rpc.EchoWallet(app_tag)
				go rpc.GetDreamsBalances(rpc.SCIDs)
				rpc.GetWalletHeight(app_tag)

				if rpc.Daemon.IsConnected() {
					rpc.Startup = false
				}

				d.SignalChannel()

			case <-d.Closing():
				logger.Printf("[%s] Closing...", app_tag)
				ticker.Stop()
				d.CloseAllDapps()
				time.Sleep(time.Second)
				done <- struct{}{}
				return
			}
		}
	}()

	connect_box := dwidget.NewHorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
	}

	connect_box.AddDaemonOptions(config.Daemon)

	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	max := container.NewMax(d.Background, LayoutAllItems(&d))

	go func() {
		time.Sleep(450 * time.Millisecond)
		w.SetContent(container.NewBorder(nil, container.NewVBox(layout.NewSpacer(), connect_box.Container), nil, nil, max))
	}()

	w.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed", app_tag)
}
