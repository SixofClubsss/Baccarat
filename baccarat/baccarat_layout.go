package baccarat

import (
	"image/color"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

var B dreams.ContainerStack

var logHand = widget.NewMultiLineEntry()

// Baccarat tab layout
func LayoutAllItems(d *dreams.AppObject) *fyne.Container {
	B.LeftLabel = widget.NewLabel("")
	B.RightLabel = widget.NewLabel("")
	B.LeftLabel.SetText("Total Hands Played: " + bacc.display.stats.total + "      Player Wins: " + bacc.display.stats.player + "      Ties: " + bacc.display.stats.ties + "      Banker Wins: " + bacc.display.stats.banker + "      Min Bet is " + bacc.display.tableMin + " dReams, Max Bet is " + bacc.display.tableMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	results := canvas.NewText("", color.White)
	results.Move(fyne.NewPos(564, 287))
	results.Alignment = fyne.TextAlignCenter

	// Waiting for block gif
	waiting_cont := container.NewVBox()

	var err error
	waiting, err = xwidget.NewAnimatedGifFromResource(ResourceLoadingGif)
	if err != nil {
		logger.Errorln("[Baccarat] Err loading gif")
	} else {
		waiting.SetMinSize(fyne.NewSize(100, 100))
		waiting.Hide()
		waiting_cont = container.NewVBox(dwidget.NewSpacer(0, 177), container.NewHBox(dwidget.NewSpacer(500, 0), waiting))
	}

	B.Back = *container.NewWithoutLayout(
		BaccTable(resourceBaccTablePng),
		results)

	B.Front = *clearBaccCards()

	logHand.Disable()

	B.DApp = container.NewStack(
		container.NewVBox(
			dwidget.LabelColor(container.NewHBox(B.LeftLabel, layout.NewSpacer(), B.RightLabel))),
		&B.Back,
		&B.Front,
		container.NewAdaptiveGrid(3,
			layout.NewSpacer(),
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), container.NewStack(container.NewBorder(dwidget.NewSpacer(250, 30), nil, nil, nil, logHand)))),
		baccaratButtons(d),
		waiting_cont)

	// Main process
	go fetch(d)

	return B.DApp
}
