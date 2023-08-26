package baccarat

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/Holdero/holdero"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
)

// Balance and swap container from dReams repo
func PlaceSwap() *container.Split {
	pair_opts := []string{"DERO-dReams", "dReams-DERO"}
	select_pair := widget.NewSelect(pair_opts, nil)
	select_pair.PlaceHolder = "Pairs"
	select_pair.SetSelectedIndex(0)

	assets := []string{}
	for asset := range rpc.Wallet.Display.Balance {
		assets = append(assets, asset)
	}

	sort.Strings(assets)

	menu.Assets.Balances = widget.NewList(
		func() int {
			return len(assets)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(assets[i] + fmt.Sprintf(": %s", rpc.DisplayBalance(assets[i])))
		})

	balance_tabs := container.NewAppTabs(
		container.NewTabItem("Balances", container.NewBorder(nil, menu.NameEntry(), nil, nil, menu.Assets.Balances)))

	var swap_entry *dwidget.DeroAmts
	var swap_boxes *fyne.Container

	max := container.NewMax()
	swap_tabs := container.NewAppTabs()

	swap_button := widget.NewButton("Swap", nil)
	swap_button.OnTapped = func() {
		switch select_pair.Selected {
		case "DERO-dReams":
			f, err := strconv.ParseFloat(swap_entry.Text, 64)
			if err == nil && swap_entry.Validate() == nil {
				if amt := (f * 333) * 100000; amt > 0 {
					max.Objects[0] = holdero.DreamsConfirm(1, amt, max, swap_tabs)
					max.Refresh()
				}
			}
		case "dReams-DERO":
			f, err := strconv.ParseFloat(swap_entry.Text, 64)
			if err == nil && swap_entry.Validate() == nil {
				if amt := f * 100000; amt > 0 {
					max.Objects[0] = holdero.DreamsConfirm(2, amt, max, swap_tabs)
					max.Refresh()
				}
			}
		}
	}

	swap_entry, swap_boxes = menu.CreateSwapContainer(select_pair.Selected)
	menu.Assets.Swap = container.NewBorder(select_pair, swap_button, nil, nil, swap_boxes)
	menu.Assets.Swap.Hide()

	select_pair.OnChanged = func(s string) {
		split := strings.Split(s, "-")
		if len(split) != 2 {
			return
		}

		swap_entry, swap_boxes = menu.CreateSwapContainer(s)

		menu.Assets.Swap.Objects[0] = swap_boxes
		menu.Assets.Swap.Refresh()
	}

	swap_tabs = container.NewAppTabs(container.NewTabItem("Swap", container.NewCenter(menu.Assets.Swap)))
	max.Add(swap_tabs)

	full := container.NewHSplit(container.NewMax(bundle.NewAlpha120(), balance_tabs), max)
	full.SetOffset(0.66)

	return full
}
