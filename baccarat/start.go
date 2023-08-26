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
	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/Holdero/holdero"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"
)

const app_tag = "Baccarat"

// Run Baccarat as a single dApp
func StartApp() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	menu.InitLogrusLog(logrus.InfoLevel)
	config := menu.ReadDreamsConfig(app_tag)

	// Initialize Fyne app and window
	a := app.NewWithID(fmt.Sprintf("%s Desktop Client", app_tag))
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.SetIcon(holdero.ResourcePokerBotIconPng)
	w.Resize(fyne.NewSize(1400, 800))
	w.SetMaster()
	done := make(chan struct{})

	// Initialize dReams AppObject and close func
	dreams.Theme.Img = *canvas.NewImageFromResource(nil)
	d := dreams.AppObject{
		App:        a,
		Window:     w,
		Background: container.NewMax(&dreams.Theme.Img),
	}
	d.SetChannels(1)

	closeFunc := func() {
		save := dreams.SaveData{
			Skin:   config.Skin,
			DBtype: menu.Gnomes.DBType,
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.WriteDreamsConfig(save)
		menu.CloseAppSignal(true)
		menu.Gnomes.Stop(app_tag)
		d.StopProcess()
		w.Close()
	}

	w.SetCloseIntercept(closeFunc)

	// Handle ctrl-c close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		closeFunc()
	}()

	// Initialize vars
	rpc.InitBalances()

	// Stand alone process
	go func() {
		time.Sleep(3 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
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

				if rpc.Wallet.IsConnected() {
					menu.Assets.Swap.Show()
					menu.Assets.Balances.Refresh()
				} else {
					menu.Assets.Swap.Hide()
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

	// Create dwidget connection box with controls
	connect_box := dwidget.NewHorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
	}

	connect_box.AddDaemonOptions(config.Daemon)
	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	// Initialize asset widgets
	asset_selects := []fyne.Widget{
		menu.NameEntry().(*fyne.Container).Objects[1].(*widget.Select),
		holdero.FaceSelect(),
		holdero.BackSelect(),
		dreams.ThemeSelect(),
		holdero.AvatarSelect(menu.Assets.Asset_map),
	}

	// Layout tabs
	tabs := container.NewAppTabs(
		container.NewTabItem(app_tag, LayoutAllItems(&d)),
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, asset_selects, holdero.ResourcePokerBotIconPng, d.Window)),
		container.NewTabItem("Swap", PlaceSwap()),
		container.NewTabItem("Log", rpc.SessionLog()))

	tabs.SetTabLocation(container.TabLocationBottom)

	go func() {
		time.Sleep(450 * time.Millisecond)
		w.SetContent(container.NewMax(d.Background, container.NewMax(bundle.NewAlpha180(), tabs), container.NewVBox(layout.NewSpacer(), connect_box.Container)))
	}()

	w.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed", app_tag)
}
