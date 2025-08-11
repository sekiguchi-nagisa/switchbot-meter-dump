package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alecthomas/kong"
	"github.com/go-ble/ble"
)

var CLI struct {
	Version kong.VersionFlag `short:"v" help:"Show version information"`
	Addr    string           `short:"a" required:"" help:"Set target SwitchBot Meter address"`
	Output  string           `short:"o" required:"" help:"Set output file"`
	Timeout time.Duration    `short:"t" default:"5s" help:"Set scan timeout"`
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

	if CLI.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	scanner, err := NewSwitchBotScanner()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// scan for SwitchBot devices
	slog.Debug("scan for SwitchBot devices")
	devices, err := scanner.ScanForSwitchBotDevices(ble.NewAddr(CLI.Addr), CLI.Timeout)
	if err != nil {
		slog.Error(fmt.Sprintf("device scan error: %v", err))
		os.Exit(1)
	}

	if len(devices) == 0 {
		slog.Error("SwitchBot device not found")
		os.Exit(1)
	}

	// parse manufacturer data
	device := devices[0]
	if slog.Default().Enabled(nil, slog.LevelDebug) {
		slog.Debug("device: " + device.Addr().String())
		for _, data := range device.ServiceData() {
			slog.Debug(fmt.Sprintf("service data: %s, %v", data.UUID, data.Data))
		}
		slog.Debug(fmt.Sprintf("manufacturer data: %v", device.ManufacturerData()))
	}

	meterData, err := DecodeManufacturerData(device.ManufacturerData())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info(meterData.String())

	// save to DB
	db, err := sql.Open("sqlite3", CLI.Output)
	if err != nil {
		slog.Error(fmt.Sprintf("Error Open: %s", err.Error()))
		os.Exit(1)
	}
	defer func(conn *sql.DB) {
		_ = conn.Close()
	}(db)
	err = InsertMeter(db, meterData)
	if err != nil {
		slog.Error(fmt.Sprintf("InsertMeter failed: %s", err.Error()))
		os.Exit(1)
	}
}
