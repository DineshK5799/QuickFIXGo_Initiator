// Copyright (c) quickfixengine.org  All rights reserved.
//
// This file may be distributed under the terms of the quickfixengine.org
// license as defined by quickfixengine.org and appearing in the file
// LICENSE included in the packaging of this file.
//
// This file is provided AS IS with NO WARRANTY OF ANY KIND, INCLUDING
// THE WARRANTY OF DESIGN, MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE.
//
// See http://www.quickfixengine.org/LICENSE for licensing information.
//
// Contact ask@quickfixengine.org if any conditions of this licensing
// are not clear to you.

package internal

import (
	"bufio"
	"fmt"
	"math/rand"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/shopspring/decimal"

	"github.com/quickfixgo/quickfix"

	"os"
	"strconv"
	"strings"

	fix44mdr "github.com/quickfixgo/fix44/marketdatarequest"

	fix44slr "github.com/quickfixgo/fix44/securitylistrequest"
)

func queryString(fieldName string) string {
	fmt.Printf("%v: ", fieldName)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return scanner.Text()
}

func queryDecimal(fieldName string) decimal.Decimal {
	val, err := decimal.NewFromString(queryString(fieldName))
	if err != nil {
		panic(err)
	}

	return val
}

func queryFieldChoices(fieldName string, choices []string, values []string) string {
	for i, choice := range choices {
		fmt.Printf("%v) %v\n", i+1, choice)
	}

	choiceStr := queryString(fieldName)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(choices) {
		panic(fmt.Errorf("Invalid %v: %v", fieldName, choice))
	}

	if values == nil {
		return choiceStr
	}

	return values[choice-1]
}

func QueryAction() (string, error) {
	fmt.Println()
	fmt.Println("1) Request Market Data")
	fmt.Println("2) Request Security List")
	fmt.Println("5) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

func queryVersion() (string, error) {
	fmt.Println()
	fmt.Print("Using FIX.4.4: ")
	fmt.Println()
	return quickfix.BeginStringFIX44, nil

}

func queryClOrdID() field.ClOrdIDField {
	return field.NewClOrdID(queryString("ClOrdID"))
}

func queryOrigClOrdID() field.OrigClOrdIDField {
	return field.NewOrigClOrdID(("OrigClOrdID"))
}

func querySymbol() field.SymbolField {
	return field.NewSymbol(queryString("Symbol"))
}

func querySide() field.SideField {
	choices := []string{
		"Buy",
		"Sell",
		"Sell Short",
		"Sell Short Exempt",
		"Cross",
		"Cross Short",
		"Cross Short Exempt",
	}

	values := []string{
		string(enum.Side_BUY),
		string(enum.Side_SELL),
		string(enum.Side_SELL_SHORT),
		string(enum.Side_SELL_SHORT_EXEMPT),
		string(enum.Side_CROSS),
		string(enum.Side_CROSS_SHORT),
		"A",
	}

	return field.NewSide(enum.Side(queryFieldChoices("Side", choices, values)))
}

func queryOrdType(f *field.OrdTypeField) field.OrdTypeField {
	choices := []string{
		"Market",
		"Limit",
		"Stop",
		"Stop Limit",
	}

	values := []string{
		string(enum.OrdType_MARKET),
		string(enum.OrdType_LIMIT),
		string(enum.OrdType_STOP),
		string(enum.OrdType_STOP_LIMIT),
	}

	f.FIXString = quickfix.FIXString(queryFieldChoices("OrdType", choices, values))
	return *f
}

func queryTimeInForce() field.TimeInForceField {
	choices := []string{
		"Day",
		"IOC",
		"OPG",
		"GTC",
		"GTX",
	}
	values := []string{
		string(enum.TimeInForce_DAY),
		string(enum.TimeInForce_IMMEDIATE_OR_CANCEL),
		string(enum.TimeInForce_AT_THE_OPENING),
		string(enum.TimeInForce_GOOD_TILL_CANCEL),
		string(enum.TimeInForce_GOOD_TILL_CROSSING),
	}

	return field.NewTimeInForce(enum.TimeInForce(queryFieldChoices("TimeInForce", choices, values)))
}

func queryOrderQty() field.OrderQtyField {
	return field.NewOrderQty(queryDecimal("OrderQty"), 2)
}

func queryPrice() field.PriceField {
	return field.NewPrice(queryDecimal("Price"), 2)
}

func queryStopPx() field.StopPxField {
	return field.NewStopPx(queryDecimal("Stop Price"), 2)
}

func querySenderCompID() field.SenderCompIDField {
	return field.NewSenderCompID(queryString("SenderCompID"))
}

func queryTargetCompID() field.TargetCompIDField {
	return field.NewTargetCompID(queryString("TargetCompID"))
}

func queryTargetSubID() field.TargetSubIDField {
	return field.NewTargetSubID(queryString("TargetSubID"))
}

func queryConfirm(prompt string) bool {
	fmt.Println()
	fmt.Printf("%v?: ", prompt)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	return strings.ToUpper(scanner.Text()) == "Y"
}

type header interface {
	Set(f quickfix.FieldWriter) *quickfix.FieldMap
}

func queryHeader(h header) {
	h.Set(querySenderCompID())
	h.Set(queryTargetCompID())
	if ok := queryConfirm("Use a TargetSubID"); !ok {
		return
	}

	h.Set(queryTargetSubID())
}

func queryMarketDataRequest44() fix44mdr.MarketDataRequest {
	request := fix44mdr.New(field.NewMDReqID("MARKETDATAID"),
		field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT),
		field.NewMarketDepth(0),
	)
	entryTypes := fix44mdr.NewNoMDEntryTypesRepeatingGroup()
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_BID)
	request.SetNoMDEntryTypes(entryTypes)

	// const (
	// 	SubscriptionRequestType_SNAPSHOT                                      SubscriptionRequestType = "0"
	// 	SubscriptionRequestType_SNAPSHOT_PLUS_UPDATES                         SubscriptionRequestType = "1"
	// 	SubscriptionRequestType_DISABLE_PREVIOUS_SNAPSHOT_PLUS_UPDATE_REQUEST SubscriptionRequestType = "2"
	// )

	request.SetSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT_PLUS_UPDATES)
	// request.SetSubscriptionRequestType(enum.SubscriptionRequestType_DISABLE_PREVIOUS_SNAPSHOT_PLUS_UPDATE_REQUEST)
	relatedSym := fix44mdr.NewNoRelatedSymRepeatingGroup()
	isins := []string{"SGXF48097749",
		"SG6PE4000001",
		"SG6TC3000008",
		"SG7BB1000008",
		"CH0482172324",
		"XS2357239057",
		"XS1679216801",
		"US251525AX97",
		"USF1067PAB25",
		"XS2351242461",
		"XS2201954067",
		"XS1513776374",
		"USG9T27HAA24",
		"NO0011128316",
		"FR0011606169",
		"XS2348280962",
		"US03938LBC72",
		"XS1410341389",
		"XS2627125672",
		"XS2611617700",
		"XS2611617619",
		"US86562MDG24",
		"USY72570AL17",
		"US44891CCZ41",
		"XS2787854673",
		"XS2502879096",
		"XS2022434364",
		"XS2775732451",
		"XS2775699577",
		"XS2774954577",
		"USY4841M6A22",
		"USY3815NBH36",
		"US96122QAC78",
		"HK0000963279",
		"US91282CHP95",
		"US912810TP30",
		"US912828UN88",
		"US912796YM59",
		"US91282CGT27",
		"XS2690013052",
		"US302154DZ91",
		"USQ82780AG49",
		"XS2675743160",
		"USJ54675BC69",
		"SGXF24733614"}
	rand.Seed(time.Now().Unix())
	symbol := isins[rand.Intn(len(isins))]
	fmt.Println(`Fetching market data for isin------------------> `, symbol)
	// relatedSym.Add().SetSymbol(symbol)
	relatedSym.Add().SetSymbol("US00084EAE86")
	request.SetNoRelatedSym(relatedSym)
	queryHeader(request.Header)
	return request
}

func QueryMarketDataRequest() error {
	req := queryMarketDataRequest44()
	if queryConfirm("Send MarketDataRequest") {
		return quickfix.Send(req)
	}
	return nil
}

func querySecurityListRequest44() fix44slr.SecurityListRequest {
	request := fix44slr.New(field.NewSecurityReqID("SECURITYREQID"), field.NewSecurityListRequestType(enum.SecurityListRequestType_ALL_SECURITIES))
	queryHeader(request.Header)
	return request
}

func QuerySecurityListRequest() error {
	beginString, err := queryVersion()
	if err != nil {
		return err
	}

	var req quickfix.Messagable
	switch beginString {
	case quickfix.BeginStringFIX44:
		req = querySecurityListRequest44()

	default:
		return fmt.Errorf("No test for version %v", beginString)
	}

	if queryConfirm("Send Seucrity List Request") {
		return quickfix.Send(req)
	}

	return nil

}
