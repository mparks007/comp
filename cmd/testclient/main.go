package main

import (
	"flag"
	"fmt"

	"strings"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"strconv"

	"github.concur.com/mparks/adder/circuit"
)

var actionType = flag.String("action", "", "Type of action to take (e.g. battery/switch/relay/and/or/nand/nor/xor/invert/osc/select/halfadd/fulladd/add/3add/sub/comp/flipflop/levellatch/nbitlatch/edgelatch/freqdiv)")
var aIn = flag.String("aIn", "0", "Power indicator at pin A (1 or 0)")
var bIn = flag.String("bIn", "0", "Power indicator at pin B (1 or 0)")
var cIn = flag.String("cIn", "0", "Power indicator at pin C (1 or 0)")
var bitString1 = flag.String("bits1", "00000000", "First string of bits in an action (e.g. 11110000)")
var bitString2 = flag.String("bits2", "00000000", "Second string of bits in an action that takes two inputs (e.g. 00001111)")

func main() {
	flag.Parse()

	_, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		fmt.Println("Error:" + err.Error())
		return
	}

	execute()
}

func execute() {

	switch *actionType {
	case "battery":

		fmt.Println("Creating battery")
		batt := circuit.NewBattery()
		batt.WireUp(func(state bool) {
			fmt.Printf("Is Charged?: %v\n", batt.GetIsPowered())
		})
		fmt.Println("Discharging battery")
		batt.Discharge()
		fmt.Println("Recharging battery")
		batt.Charge()
		fmt.Println("Recharging charged battery (no affect)")
		batt.Charge()

	case "switch":

		state, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state: " + *aIn)
			return
		}

		fmt.Println("Creating switch")
		sw := circuit.NewSwitch(state)
		sw.WireUp(func(state bool) {
			fmt.Printf("Is On?: %v\n", sw.GetIsPowered())
		})
		if sw.GetIsPowered() {
			fmt.Println("Turning switch off")
			sw.Set(false)
		}

		fmt.Println("Turning switch on")
		sw.Set(true)
		fmt.Println("Turning switch on again (no affect)")
		sw.Set(true)

	case "relay":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state: " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state: " + *bIn)
			return
		}

		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)
		relay := circuit.NewRelay(sw1, sw2)

		fmt.Printf("Open circuit power?: %v\n", relay.OpenOut.GetIsPowered())
		fmt.Printf("Closed circuit power?: %v\n", relay.ClosedOut.GetIsPowered())

	case "add":
		switches1, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		switches2, err := circuit.NewNSwitchBank(*bitString2)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		addr, err := circuit.NewNBitAdder(switches1.AsPwrEmitters(), switches2.AsPwrEmitters(), nil)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		var carry string
		if addr.CarryOutAsBool() {
			carry = "1"
		}
		fmt.Printf("  %s\n+ %s\n=%1s%s\n\n", *bitString1, *bitString2, carry, addr.AsAnswerString())
	case "sub":
		switches1, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		switches2, err := circuit.NewNSwitchBank(*bitString2)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		subtr, err := circuit.NewNBitSubtractor(switches1.AsPwrEmitters(), switches2.AsPwrEmitters())
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		fmt.Printf("  %s\n- %s\n= %s\n\n", *bitString1, *bitString2, subtr.AsAnswerString())
	case "comp":
		switches1, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		c := circuit.NewOnesComplementer(switches1.AsPwrEmitters(), circuit.NewBattery())
		fmt.Println("   Input String: " + *bitString1)
		fmt.Println("Ones Complement: " + c.AsComplementString())

		switches, _ := circuit.NewNSwitchBank(strings.Repeat("0", len(c.Complements)-1) + "1")
		twosC, err := circuit.NewNBitAdder(c.Complements, switches.AsPwrEmitters(), nil)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		} else {
			fmt.Println("Twos Complement: " + twosC.AsAnswerString())
		}

	}
}
