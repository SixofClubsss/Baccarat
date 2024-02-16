package baccarat

import (
	"fmt"
	"strconv"

	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/cryptography/crypto"
	dero "github.com/deroproject/derohe/rpc"
)

type cards struct {
	card1 int
	card2 int
	card3 int
}

type baccValues struct {
	player   cards
	banker   cards
	cHeight  int
	minBet   float64
	maxBet   float64
	assetID  string
	contract string
	last     int
	found    bool
	wait     bool
	display  struct {
		tableMax string
		tableMin string
		result   string
		stats    struct {
			total  string
			player string
			banker string
			ties   string
		}
	}
}

var bacc baccValues

// Get Baccarat SC data
func fetchBaccSC() {
	if rpc.Daemon.IsConnected() && rpc.Wallet.Height > bacc.last {
		client, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      bacc.contract,
			Code:      false,
			Variables: true,
		}

		if err := client.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			logger.Errorln("[FetchBaccSC]", err)
			return
		}

		Asset_jv := result.VariableStringKeys["tokenSCID"]
		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		Player_jv := result.VariableStringKeys["Player Wins:"]
		Banker_jv := result.VariableStringKeys["Banker Wins:"]
		Min_jv := result.VariableStringKeys["Min Bet:"]
		Max_jv := result.VariableStringKeys["Max Bet:"]
		Ties_jv := result.VariableStringKeys["Ties:"]
		// dReams_jv := result.Balances[rpc.DreamsSCID]
		// Dero_jv := result.Balances[crypto.ZEROHASH.String()]
		if Asset_jv != nil {
			bacc.assetID = fmt.Sprint(Asset_jv)
		}

		if Total_jv != nil {
			bacc.display.stats.total = fmt.Sprint(Total_jv)
		}

		if Player_jv != nil {
			bacc.display.stats.player = fmt.Sprint(Player_jv)
		}

		if Banker_jv != nil {
			bacc.display.stats.banker = fmt.Sprint(Banker_jv)
		}

		if Ties_jv != nil {
			bacc.display.stats.ties = fmt.Sprint(Ties_jv)
		}

		if max, ok := Max_jv.(float64); ok {
			bacc.display.tableMax = fmt.Sprintf("%.0f", max/100000)
			bacc.maxBet = max / 100000
		} else {
			bacc.display.tableMax = "250"
			bacc.maxBet = 250
		}

		if min, ok := Min_jv.(float64); ok {
			bacc.display.tableMin = fmt.Sprintf("%.0f", min/100000)
			bacc.minBet = min / 100000
		} else {
			bacc.display.tableMin = "10"
			bacc.minBet = 10
		}

		// Update TX hand log
		bacc.last = rpc.Wallet.Height

		display, ok := result.VariableStringKeys["display"].(float64)
		if !ok {
			display = 33
		}

		total := rpc.IntType(Total_jv)
		disp := int(display)
		if total < disp {
			disp = total
		}

		var results string
		for i := total; i > total-disp; i-- {
			w := strconv.Itoa(i)
			if _, ok := result.VariableStringKeys[w+"-Hand#TXID:"].(string); ok {
				PTotal_jv := result.VariableStringKeys[w+"-Player total:"]
				BTotal_jv := result.VariableStringKeys[w+"-Banker total:"]

				p := rpc.IntType(PTotal_jv)
				b := rpc.IntType(BTotal_jv)
				if p == b {
					results = results + fmt.Sprintf("#%s [Tie], %d & %d\n", w, p, b)
				} else if p > b {
					results = results + fmt.Sprintf("#%s [Player Wins], %d over %d\n", w, p, b)
				} else {
					results = results + fmt.Sprintf("#%s [Banker Wins], %d over %d\n", w, b, p)
				}
			}
		}

		logHand.SetText(results)
	}
}

// Get Baccarat hand by TXID
func FetchHand(tx string) {
	if rpc.Daemon.IsConnected() && tx != "" {
		client, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      bacc.contract,
			Code:      false,
			Variables: true,
		}

		if err := client.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			logger.Errorln("[FetchHand]", err)
			return
		}

		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		if Total_jv != nil {
			Display_jv := result.VariableStringKeys["display"]
			start := rpc.IntType(Total_jv) - rpc.IntType(Display_jv)

			for i := start; i < start+45; i++ {
				h := "-Hand#TXID:"
				w := strconv.Itoa(i)

				if txid, ok := result.VariableStringKeys[w+h].(string); ok {
					if txid == tx {
						bacc.found = true
						bacc.player.card1 = rpc.IntType(result.VariableStringKeys[w+"-Player x:"])
						bacc.player.card2 = rpc.IntType(result.VariableStringKeys[w+"-Player y:"])
						bacc.player.card3 = rpc.IntType(result.VariableStringKeys[w+"-Player z:"])
						bacc.banker.card1 = rpc.IntType(result.VariableStringKeys[w+"-Banker x:"])
						bacc.banker.card2 = rpc.IntType(result.VariableStringKeys[w+"-Banker y:"])
						bacc.banker.card3 = rpc.IntType(result.VariableStringKeys[w+"-Banker z:"])

						p := rpc.IntType(result.VariableStringKeys[w+"-Player total:"])
						b := rpc.IntType(result.VariableStringKeys[w+"-Banker total:"])

						if p == b {
							bacc.display.result = fmt.Sprintf("Hand# %s Tie, %d & %d", w, p, b)
						} else if p > b {
							bacc.display.result = fmt.Sprintf("Hand# %s Player Wins, %d over %d", w, p, b)
						} else {
							bacc.display.result = fmt.Sprintf("Hand# %s Banker Wins, %d over %d", w, b, p)
						}

						return
					}
				}
			}
		}
	}
}

// Get last hand played
func FetchLastHand() (found bool) {
	if rpc.Daemon.IsConnected() {
		client, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      bacc.contract,
			Code:      false,
			Variables: true,
		}

		if err := client.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			logger.Errorln("[FetchLastHand]", err)
			return
		}

		Total_jv := result.VariableStringKeys["TotalHandsPlayed:"]
		if rpc.IntType(Total_jv) > 0 {
			total := fmt.Sprintf("%d", rpc.IntType(Total_jv))
			if _, ok := result.VariableStringKeys[total+"-Hand#TXID:"].(string); ok {
				found = true
				bacc.player.card1 = rpc.IntType(result.VariableStringKeys[total+"-Player x:"])
				bacc.player.card2 = rpc.IntType(result.VariableStringKeys[total+"-Player y:"])
				bacc.player.card3 = rpc.IntType(result.VariableStringKeys[total+"-Player z:"])
				bacc.banker.card1 = rpc.IntType(result.VariableStringKeys[total+"-Banker x:"])
				bacc.banker.card2 = rpc.IntType(result.VariableStringKeys[total+"-Banker y:"])
				bacc.banker.card3 = rpc.IntType(result.VariableStringKeys[total+"-Banker z:"])

				p := rpc.IntType(result.VariableStringKeys[total+"-Player total:"])
				b := rpc.IntType(result.VariableStringKeys[total+"-Banker total:"])

				if p == b {
					bacc.display.result = fmt.Sprintf("Hand# %s Tie, %d & %d", total, p, b)
				} else if p > b {
					bacc.display.result = fmt.Sprintf("Hand# %s Player Wins, %d over %d", total, p, b)
				} else {
					bacc.display.result = fmt.Sprintf("Hand# %s Banker Wins, %d over %d", total, b, p)
				}
			}
		}
	}

	return
}

// Place Baccarat bet
//   - amt to bet
//   - w defines where bet is placed (player, banker or tie)
func BaccBet(amt, w string) (tx string) {
	if bacc.assetID == "" || len(bacc.assetID) != 64 {
		rpc.PrintError("[Baccarat] Asset ID error")
		return "ID error"
	}

	client, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "PlayBaccarat"}
	arg2 := dero.Argument{Name: "betOn", DataType: "S", Value: w}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		SCID:        crypto.HashHexToHash(bacc.assetID),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        rpc.ToAtomic(amt, 1),
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(bacc.contract, "[Baccarat]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     bacc.contract,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := client.CallFor(ctx, &txid, "transfer", params); err != nil {
		rpc.PrintError("[Baccarat] Bet: %s", err)
		return
	}

	if w == "player" {
		rpc.PrintLog("[Baccarat] Player TX: %s", txid)
	} else if w == "banker" {
		rpc.PrintLog("[Baccarat] Banker TX: %s", txid)
	} else {
		rpc.PrintLog("[Baccarat] Tie TX: %s", txid)
	}

	bacc.cHeight = rpc.Wallet.Height

	return txid.TXID
}
