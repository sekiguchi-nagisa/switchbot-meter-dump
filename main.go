package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-ble/ble"
)

var CLI struct {
	Version kong.VersionFlag `short:"v" help:"Show version information"`
	Addr    string           `short:"a" required:"" help:"Set target SwitchBot Meter address"`
	Output  string           `short:"o" required:"" help:"Set output file"`
	Debug   bool             `optional:"" default:"false" help:"Set debug mode"`
}

var version = "" // for version embedding (specified like "-X main.version=v0.1.0")

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		rev := "unknown"
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				rev = setting.Value
				break
			}
		}
		var v = info.Main.Version
		if version != "" { // set by "-X main.version=v0.1.0"
			v = version
		}
		return fmt.Sprintf("%s (%s)", v, rev)
	} else {
		return "(unknown)"
	}
}

func main() {
	kong.Parse(&CLI, kong.UsageOnError(), kong.Vars{"version": getVersion()})
	if CLI.Version {
		fmt.Println(getVersion())
		os.Exit(0)
	}

	scanner, err := NewSwitchBotScanner()
	if err != nil {
		log.Fatal(err)
	}

	// scan for SwitchBot devices
	if CLI.Debug {
		fmt.Println("scan for SwitchBot devices")
	}
	devices, err := scanner.ScanForSwitchBotDevices(ble.NewAddr(CLI.Addr), 5*time.Second)
	if err != nil {
		log.Fatal("device scan error:", err)
	}

	if len(devices) == 0 {
		log.Fatal("SwitchBot device not found")
	}

	// parse manufacturer data
	device := devices[0]
	if CLI.Debug {
		fmt.Printf("device: %s\n", device.Addr())
		for _, data := range device.ServiceData() {
			fmt.Printf("service data: %s, %v\n", data.UUID, data.Data)
		}
		fmt.Printf("manufacturer data: %v\n", device.ManufacturerData())
	}

	meterData, err := ParseManufacturerData(device.ManufacturerData())
	if err != nil {
		log.Fatal("parse manufacturer data error:", err)
	}
	log.Println(meterData)
}
