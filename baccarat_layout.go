package baccarat

import (
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var B dreams.DreamsItems

// Baccarat tab layout
func LayoutAllItems(d dreams.DreamsObject) *fyne.Container {
	B.LeftLabel = widget.NewLabel("")
	B.RightLabel = widget.NewLabel("")
	B.LeftLabel.SetText("Total Hands Played: " + Display.Total_w + "      Player Wins: " + Display.Player_w + "      Ties: " + Display.Ties + "      Banker Wins: " + Display.Banker_w + "      Min Bet is " + Display.BaccMin + " dReams, Max Bet is " + Display.BaccMax)
	B.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	B.Back = *container.NewWithoutLayout(
		BaccTable(resourceBaccTablePng),
		baccResult(Display.BaccRes))

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
