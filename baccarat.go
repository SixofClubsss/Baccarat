package baccarat

import (
	"encoding/hex"
	"encoding/json"
	"image/color"
	"math/rand"
	"strconv"
	"time"

	holdero "github.com/SixofClubsss/Holdero"
	"github.com/civilware/Gnomon/structures"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Tables map[string]string
var logger = structures.Logger.WithFields(logrus.Fields{})

func DreamsMenuIntro() (entries map[string][]string) {
	entries = map[string][]string{
		"Baccarat": {
			"A popular table game, where closest to 9 wins",
			"Bet on player, banker or tie as the winning outcome",
			"Select table with bottom left drop down to choose currency"},
	}

	return
}

// Function for when Baccarat tab is selected
func OnTabSelected(d *dreams.DreamsObject) {
	GetBaccTables()
	BaccRefresh(d)
	if rpc.Wallet.IsConnected() && Bacc.Display {
		ActionBuffer(false)
	}
}

// Main Baccarat process
func fetch(d *dreams.DreamsObject) {
	Bacc.Display = true
	time.Sleep(3 * time.Second)
	for {
		select {
		case <-d.Receive():
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				disableBaccActions(true)
				BaccRefresh(d)
				d.WorkDone()
				continue
			}

			fetchBaccSC()
			BaccRefresh(d)
			d.WorkDone()
		case <-d.CloseDapp():
			logger.Println("[Baccarat] Done")
			return
		}
	}
}

// Baccarat object buffer when action triggered
func ActionBuffer(d bool) {
	if d {
		B.Actions.Hide()
		Bacc.P_card1 = 99
		Bacc.P_card2 = 99
		Bacc.P_card3 = 99
		Bacc.B_card1 = 99
		Bacc.B_card2 = 99
		Bacc.B_card3 = 99
		Bacc.Last = ""
		Display.BaccRes = "Wait for Block..."
	} else {
		if rpc.Daemon.IsConnected() && Display.BaccRes != "Wait for Block..." {
			B.Actions.Show()
		}
	}

	B.Actions.Refresh()
}

// Disable Baccarat actions
func disableBaccActions(d bool) {
	if d {
		B.Actions.Hide()
	} else {
		B.Actions.Show()
	}

	B.Actions.Refresh()
}

// Baccarat hand result display label
func baccResult(r string) *canvas.Text {
	label := canvas.NewText(r, color.White)
	label.Move(fyne.NewPos(564, 237))
	label.Alignment = fyne.TextAlignCenter

	return label
}

// Baccarat action objects
func baccaratButtons(w fyne.Window) fyne.CanvasObject {
	entry := dwidget.NewDeroEntry("", 1, 0)
	entry.PlaceHolder = "dReams:"
	entry.AllowFloat = false
	entry.SetText("10")
	entry.Validator = validation.NewRegexp(`^\d{1,}$`, "Int required")
	entry.OnChanged = func(s string) {
		if rpc.Daemon.IsConnected() {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if f < Bacc.MinBet {
					entry.SetText(Display.BaccMin)
				}

				if f > Bacc.MaxBet {
					entry.SetText(Display.BaccMax)
				}
			}

			if entry.Validate() != nil {
				entry.SetText(Display.BaccMin)
			}
		}
	}

	player_button := widget.NewButton("Player", func() {
		ActionBuffer(true)
		Bacc.Found = false
		Bacc.Display = false
		if tx := BaccBet(entry.Text, "player"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	banker_button := widget.NewButton("Banker", func() {
		ActionBuffer(true)
		Bacc.Found = false
		Bacc.Display = false
		if tx := BaccBet(entry.Text, "banker"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	tie_button := widget.NewButton("Tie", func() {
		ActionBuffer(true)
		Bacc.Found = false
		Bacc.Display = false
		if tx := BaccBet(entry.Text, "tie"); tx == "ID error" {
			dialog.NewInformation("Baccarat", "Select a table", w).Show()
		}
	})

	amt_box := container.NewHScroll(entry)
	amt_box.SetMinSize(fyne.NewSize(100, 40))

	actions := container.NewVBox(
		player_button,
		banker_button,
		tie_button,
		amt_box)

	var searched string
	search_entry := widget.NewEntry()
	search_entry.SetPlaceHolder("TXID:")
	search_button := widget.NewButton("     Search    ", func() {
		txid := search_entry.Text
		if len(txid) == 64 && txid != searched {
			searched = txid
			ActionBuffer(true)
			Display.BaccRes = "Searching..."
			Bacc.Found = false
			Bacc.Display = false
			FetchBaccHand(txid)
			if !Bacc.Found {
				Display.BaccRes = "Hand Not Found"
				ActionBuffer(false)
			}
		}
	})

	Display.BaccMin = "10"
	table_opts := []string{"dReams"}
	table_select := widget.NewSelect(table_opts, func(s string) {
		switch s {
		case "dReams":
			Bacc.Contract = rpc.BaccSCID
		default:
			Bacc.Contract = Tables[s]
		}
		fetchBaccSC()
		entry.SetText(Display.BaccMin)
	})
	table_select.PlaceHolder = "Select Table:"
	table_select.SetSelectedIndex(0)

	search := container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2,
			table_select,
			container.NewBorder(nil, nil, nil, search_button, search_entry)))

	B.Actions = *container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), actions),
		search)

	B.Actions.Hide()

	return &B.Actions
}

// Baccarat table image
func BaccTable(img fyne.Resource) fyne.CanvasObject {
	table_img := canvas.NewImageFromResource(img)
	table_img.Resize(fyne.NewSize(1100, 600))
	table_img.Move(fyne.NewPos(5, 0))

	return table_img
}

// Gets list of current Baccarat tables from on chain store and refresh options
func GetBaccTables() {
	if rpc.Daemon.IsConnected() {
		Tables = make(map[string]string)
		if table_map, ok := rpc.FindStringKey(rpc.RatingSCID, "bacc_tables", rpc.Daemon.Rpc).(string); ok {
			if str, err := hex.DecodeString(table_map); err == nil {
				json.Unmarshal([]byte(str), &Tables)
			}
		}

		table_names := make([]string, 0, len(Tables))
		for name := range Tables {
			table_names = append(table_names, name)
		}

		table_select := B.Actions.Objects[2].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Select)
		table_select.Options = []string{"dReams"}
		table_select.Options = append(table_select.Options, table_names...)
		table_select.Refresh()
	}
}

// Set Baccarat player card images
func playerCards(c1, c2, c3 int) fyne.CanvasObject {
	card1 := holdero.DisplayCard(c1)
	card2 := holdero.DisplayCard(c2)
	card3 := holdero.DisplayCard(c3)

	card1.Resize(fyne.NewSize(110, 150))
	card1.Move(fyne.NewPos(180, 39))

	card2.Resize(fyne.NewSize(110, 150))
	card2.Move(fyne.NewPos(290, 39))

	card3.Resize(fyne.NewSize(110, 150))
	card3.Move(fyne.NewPos(400, 39))

	return container.NewWithoutLayout(card1, card2, card3)
}

// Set Baccarat banker card images
func bankerCards(c1, c2, c3 int) fyne.CanvasObject {
	card1 := holdero.DisplayCard(c1)
	card2 := holdero.DisplayCard(c2)
	card3 := holdero.DisplayCard(c3)

	card1.Resize(fyne.NewSize(110, 150))
	card1.Move(fyne.NewPos(600, 39))

	card2.Resize(fyne.NewSize(110, 150))
	card2.Move(fyne.NewPos(710, 39))

	card3.Resize(fyne.NewSize(110, 150))
	card3.Move(fyne.NewPos(820, 39))

	return container.NewWithoutLayout(card1, card2, card3)
}

// Place and refresh Baccarat card images
func showBaccCards() *fyne.Container {
	var drawP, drawB int
	if Bacc.P_card3 == 0 {
		drawP = 99
	} else {
		drawP = Bacc.P_card3
	}

	if Bacc.B_card3 == 0 {
		drawB = 99
	} else {
		drawB = Bacc.B_card3
	}

	content := *container.NewWithoutLayout(
		playerCards(baccSuit(Bacc.P_card1), baccSuit(Bacc.P_card2), baccSuit(drawP)),
		bankerCards(baccSuit(Bacc.B_card1), baccSuit(Bacc.B_card2), baccSuit(drawB)))

	Bacc.Display = true
	ActionBuffer(false)

	return &content
}

func clearBaccCards() *fyne.Container {
	content := *container.NewWithoutLayout(
		playerCards(99, 99, 99),
		bankerCards(99, 99, 99))

	return &content
}

// Refresh all Baccarat objects
func BaccRefresh(d *dreams.DreamsObject) {
	asset_name := rpc.GetAssetSCIDName(Bacc.AssetID)
	B.LeftLabel.SetText("Total Hands Played: " + Display.Total_w + "      Player Wins: " + Display.Player_w + "      Ties: " + Display.Ties + "      Banker Wins: " + Display.Banker_w + "      Min Bet is " + Display.BaccMin + " " + asset_name + ", Max Bet is " + Display.BaccMax)
	B.RightLabel.SetText(asset_name + " Balance: " + rpc.DisplayBalance(asset_name) + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	if !Bacc.Display {
		B.Front.Objects[0] = clearBaccCards()
		FetchBaccHand(Bacc.Last)
		if Bacc.Found {
			B.Front.Objects[0] = showBaccCards()
		}
		B.Front.Objects[0].Refresh()
	}

	if rpc.Wallet.Height > Bacc.CHeight+3 && !Bacc.Found {
		Display.BaccRes = ""
		ActionBuffer(false)
	}

	B.Back.Objects[1].(*canvas.Text).Text = Display.BaccRes
	B.Back.Objects[1].Refresh()

	B.DApp.Refresh()

	if Bacc.Found && !Bacc.Notified {
		if !d.IsWindows() {
			Bacc.Notified = d.Notification("dReams - Baccarat", Display.BaccRes)
		}
	}
}

// Create a random suit for baccarat card
func baccSuit(card int) (suited int) {
	if card == 99 {
		return 99
	}

	seed := rand.NewSource(time.Now().UnixNano())
	y := rand.New(seed)
	x := y.Intn(4) + 1

	if card == 0 {
		seed := rand.NewSource(time.Now().UnixNano())
		y := rand.New(seed)
		x := y.Intn(16) + 1

		switch x {
		case 1:
			suited = 10
		case 2:
			suited = 11
		case 3:
			suited = 12
		case 4:
			suited = 13
		case 5:
			suited = 23
		case 6:
			suited = 24
		case 7:
			suited = 25
		case 8:
			suited = 26
		case 9:
			suited = 36
		case 10:
			suited = 37
		case 11:
			suited = 38
		case 12:
			suited = 39
		case 13:
			suited = 49
		case 14:
			suited = 50
		case 15:
			suited = 51
		case 16:
			suited = 52
		}

		return
	}

	switch card {
	case 1:
		switch x {
		case 1:
			suited = 1
		case 2:
			suited = 14
		case 3:
			suited = 27
		case 4:
			suited = 40
		}
	case 2:
		switch x {
		case 1:
			suited = 2
		case 2:
			suited = 15
		case 3:
			suited = 28
		case 4:
			suited = 41
		}
	case 3:
		switch x {
		case 1:
			suited = 3
		case 2:
			suited = 16
		case 3:
			suited = 29
		case 4:
			suited = 42
		}
	case 4:
		switch x {
		case 1:
			suited = 4
		case 2:
			suited = 17
		case 3:
			suited = 30
		case 4:
			suited = 43
		}
	case 5:
		switch x {
		case 1:
			suited = 5
		case 2:
			suited = 18
		case 3:
			suited = 31
		case 4:
			suited = 44
		}
	case 6:
		switch x {
		case 1:
			suited = 6
		case 2:
			suited = 19
		case 3:
			suited = 32
		case 4:
			suited = 45
		}
	case 7:
		switch x {
		case 1:
			suited = 7
		case 2:
			suited = 20
		case 3:
			suited = 33
		case 4:
			suited = 46
		}
	case 8:
		switch x {
		case 1:
			suited = 8
		case 2:
			suited = 21
		case 3:
			suited = 34
		case 4:
			suited = 47
		}
	case 9:
		switch x {
		case 1:
			suited = 9
		case 2:
			suited = 22
		case 3:
			suited = 35
		case 4:
			suited = 48
		}
	case 10:
		switch x {
		case 1:
			suited = 10
		case 2:
			suited = 23
		case 3:
			suited = 36
		case 4:
			suited = 49
		}
	case 11:
		switch x {
		case 1:
			suited = 11
		case 2:
			suited = 24
		case 3:
			suited = 37
		case 4:
			suited = 50
		}
	case 12:
		switch x {
		case 1:
			suited = 12
		case 2:
			suited = 25
		case 3:
			suited = 38
		case 4:
			suited = 51
		}
	case 13:
		switch x {
		case 1:
			suited = 13
		case 2:
			suited = 26
		case 3:
			suited = 39
		case 4:
			suited = 52
		}
	}

	return
}
