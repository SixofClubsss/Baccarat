package baccarat

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/Holdero/holdero"
	"github.com/blang/semver/v4"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"
)

const (
	appName = "Baccarat"
	appID   = "dreamdapps.io.baccarat"
)

var version = semver.MustParse("0.3.1-dev.x")
var gnomon = gnomes.NewGnomes()

// Check baccarat package version
func Version() semver.Version {
	return version
}

// Run Baccarat as a single dApp
func StartApp() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	// Initialize logrus logger to stdout
	gnomes.InitLogrusLog(logrus.InfoLevel)

	// Read config.json file
	config := menu.GetSettings(appName)

	// Initialize Fyne app and window as dreams.AppObject
	d := dreams.NewFyneApp(
		appID,
		appName,
		"On-chain Baccarat",
		bundle.DeroTheme(config.Skin),
		holdero.ResourceCardsIconPng,
		menu.DefaultBackgroundResource(),
		true)

	// Set one channel for Baccarat routine
	d.SetChannels(1)

	// Initialize closing channels and func
	done := make(chan struct{})

	closeFunc := func() {
		save := dreams.SaveData{
			Skin:   config.Skin,
			DBtype: gnomon.DBStorageType(),
			Theme:  dreams.Theme.Name,
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.StoreSettings(save)
		menu.SetClose(true)
		gnomon.Stop(appName)
		d.StopProcess()
		d.Window.Close()
	}

	d.Window.SetCloseIntercept(closeFunc)

	// Handle ctrl-c close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		closeFunc()
	}()

	// Initialize vars

	// Stand alone process
	go func() {
		time.Sleep(3 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				rpc.Ping()
				rpc.Wallet.Sync()

				if rpc.Wallet.IsConnected() {
					menu.Assets.Swap.Show()
					menu.Assets.Balances.List.Refresh()
				} else {
					menu.Assets.Swap.Hide()
				}

				d.SignalChannel()

			case <-d.Closing():
				logger.Printf("[%s] Closing...", appName)
				ticker.Stop()
				d.CloseAllDapps()
				time.Sleep(time.Second)
				done <- struct{}{}
				return
			}
		}
	}()

	// Create dwidget connection box, using default OnTapped for RPC/XSWD connections
	connection := dwidget.NewHorizontalEntries(appName, 1, &d)

	// Set any saved daemon configs
	connection.AddDaemonOptions(config.Daemon)

	// Adding dReams indicator panel for wallet, daemon and Gnomon
	connection.AddIndicator(menu.StartIndicators(nil))

	// Initialize profile widgets
	line := canvas.NewLine(bundle.TextColor)
	form := []*widget.FormItem{}
	form = append(form, widget.NewFormItem("Name", menu.NameEntry()))
	form = append(form, widget.NewFormItem("", layout.NewSpacer()))
	form = append(form, widget.NewFormItem("", container.NewVBox(line)))
	form = append(form, widget.NewFormItem("Avatar", holdero.AvatarSelect(menu.Assets.SCIDs)))
	form = append(form, widget.NewFormItem("Theme", menu.ThemeSelect(&d)))
	form = append(form, widget.NewFormItem("Card Deck", holdero.FaceSelect(menu.Assets.SCIDs, &d)))
	form = append(form, widget.NewFormItem("Card Back", holdero.BackSelect(menu.Assets.SCIDs)))
	form = append(form, widget.NewFormItem("", layout.NewSpacer()))
	form = append(form, widget.NewFormItem("", container.NewVBox(line)))

	profile := container.NewCenter(container.NewBorder(dwidget.NewSpacer(450, 0), nil, nil, nil, widget.NewForm(form...)))

	// Layout tabs
	tabs := container.NewAppTabs(
		container.NewTabItem(appName, LayoutAll(&d)),
		container.NewTabItem("Assets", menu.PlaceAssets(appName, profile, nil, holdero.ResourceCardsCirclePng, &d)),
		container.NewTabItem("Swap", holdero.PlaceSwap(&d)),
		container.NewTabItem("Log", rpc.SessionLog(appName, version)))

	tabs.SetTabLocation(container.TabLocationBottom)

	// Start app and place layout
	go func() {
		time.Sleep(450 * time.Millisecond)
		d.Window.SetContent(container.NewStack(d.Background, container.NewStack(bundle.NewAlpha180(), tabs), container.NewVBox(layout.NewSpacer(), connection.Container)))
	}()

	d.Window.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed", appName)
}
