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
)

var B dreams.ContainerStack

// Baccarat tab layout
func LayoutAllItems(d *dreams.AppObject) *fyne.Container {
	B.LeftLabel = widget.NewLabel("")
	B.RightLabel = widget.NewLabel("")
	B.LeftLabel.SetText("Total Hands Played: " + bacc.display.stats.total + "      Player Wins: " + bacc.display.stats.player + "      Ties: " + bacc.display.stats.ties + "      Banker Wins: " + bacc.display.stats.banker + "      Min Bet is " + bacc.display.tableMin + " dReams, Max Bet is " + bacc.display.tableMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	results := canvas.NewText("", color.White)
	results.Move(fyne.NewPos(564, 237))
	results.Alignment = fyne.TextAlignCenter

	B.Back = *container.NewWithoutLayout(
		BaccTable(resourceBaccTablePng),
		results)

	B.Front = *clearBaccCards()

	bacc_label := container.NewHBox(B.LeftLabel, layout.NewSpacer(), B.RightLabel)

	B.DApp = container.NewVBox(
		dwidget.LabelColor(bacc_label),
		&B.Back,
		&B.Front,
		layout.NewSpacer(),
		baccaratButtons(d.Window))

	// Main process
	go fetch(d)

	return B.DApp
}
