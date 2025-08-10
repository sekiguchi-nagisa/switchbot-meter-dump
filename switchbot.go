package main

import (
	"context"
	"errors"
	"fmt"
	"time"
	"log"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

// SwitchBot Meter のサービスとキャラクタリスティックUUID
var SwitchBotServiceUUID = ble.MustParse("cba20d00-224d-11e6-9fb8-0002a5d5c51b")
var SwitchBotCharacteristicUUID = ble.MustParse("cba20002-224d-11e6-9fb8-0002a5d5c51b")

type MeterData struct {
	Temperature float64
	Humidity    int
	Battery     int
	Timestamp   time.Time
}

func (m MeterData) String() string {
	return fmt.Sprintf("Timestamp: %s, Temperature: %.1f°C, Humidity: %d%%, Battery: %d%%",
		m.Timestamp.Format("2006-01-02 15:04:05"), m.Temperature, m.Humidity, m.Battery)
}

type SwitchBotScanner struct {
	ctx    context.Context
	client ble.Client
}

func NewSwitchBotScanner() (*SwitchBotScanner, error) {
	d, err := dev.NewDevice("default")
	if err != nil {
		return nil, fmt.Errorf("BLE device initialization failed: %v", err)
	}
	ble.SetDefaultDevice(d)

	return &SwitchBotScanner{
		ctx: context.Background(),
	}, nil
}

func (s *SwitchBotScanner) ScanForSwitchBotDevices(addr ble.Addr, timeout time.Duration) ([]ble.Advertisement, error) {
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	var devices []ble.Advertisement

	advHandler := func(a ble.Advertisement) {
		devices = append(devices, a)
		log.Printf("found device: %s (RSSI: %d)\n", a.Addr(), a.RSSI())
	}
	advFilter := func(a ble.Advertisement) bool {
		return a.Addr().String() == addr.String()
	}

	err := ble.Scan(ctx, false, advHandler, advFilter)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return devices, nil
}

func ParseManufacturerData(data []byte) (MeterData, error) {
	// [105 9 176 233 254 87 31 201 44 100 8 153 61 0 4]
	fmt.Printf("manufacturer data: %v\n", data)
	return MeterData{}, nil
}
