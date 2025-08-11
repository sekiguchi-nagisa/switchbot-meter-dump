package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

type MeterData struct {
	Temperature float64
	Humidity    uint
	Battery     uint
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
		slog.Info(fmt.Sprintf("found device: %s (RSSI: %d)\n", a.Addr(), a.RSSI()))
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

func DecodeManufacturerData(data []byte) (MeterData, error) {
	if len(data) != 15 {
		return MeterData{}, errors.New("invalid manufacturer data length")
	}

	// [105 9 176 233 254 87 31 201 44 100 8 153 61 0 4]
	//   battery: 100
	//   humidity: 61
	//   temperature: (153 & 0x7f) + (8 & 0xf)/10.0 = 25.8
	battery := uint(data[9] & 0x7f)
	humidity := uint(data[12] & 0x7f)
	temperature := float64(data[10]&0xf)/10.0 + float64(data[11]&0x7f)
	if data[11] < 128 {
		temperature = -temperature
	}
	return MeterData{
		Temperature: temperature,
		Humidity:    humidity,
		Battery:     battery,
		Timestamp:   time.Now(),
	}, nil
}
