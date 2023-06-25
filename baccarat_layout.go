package baccarat

import (
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/dwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

var B dreams.DreamsItems

// Baccarat tab layout
func LayoutAllItems(d dreams.DreamsObject) *fyne.Container {
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
