package sgp30

import (
	"encoding/binary"
	"errors"
	"log"
	"strconv"
	"time"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/i2c"
)

const (
	InitAirQuality       uint16 = 0x2003
	MeasureAirQuality    uint16 = 0x2008
	GetIAQBaseline       uint16 = 0x2015
	SetIAQBaseline       uint16 = 0x201e
	SetHumidity          uint16 = 0x2061
	MeasureTest          uint16 = 0x2032
	GetFeatureSetVersion uint16 = 0x202f
	MeasureRawSignals    uint16 = 0x2050
	GetTVOCBaseline      uint16 = 0x20b3
	SetTVOCBaseline      uint16 = 0x2077

	I2CAddress uint16 = 0x58
)

// commandDuration maps the defined maximum measurement duration from the sensor
var commandDuration = map[uint16]time.Duration{
	InitAirQuality:       time.Millisecond * 10,
	MeasureAirQuality:    time.Millisecond * 12,
	GetIAQBaseline:       time.Millisecond * 10,
	SetIAQBaseline:       time.Millisecond * 10,
	SetHumidity:          time.Millisecond * 10,
	MeasureTest:          time.Millisecond * 220,
	GetFeatureSetVersion: time.Millisecond * 10,
	MeasureRawSignals:    time.Millisecond * 25,
	GetTVOCBaseline:      time.Millisecond * 10,
	SetTVOCBaseline:      time.Millisecond * 10,
}

// commandResponseLength maps the defined response length including the CRC
var commandResponseLength = map[uint16]int{
	MeasureAirQuality:    6,
	GetIAQBaseline:       6,
	MeasureTest:          3,
	GetFeatureSetVersion: 3,
	MeasureRawSignals:    6,
	GetTVOCBaseline:      3,
}

// CO2 represents the current carbon dioxide value in ppm
type CO2 uint16

func (c CO2) String() string {
	return strconv.Itoa(int(c)) + "ppm"
}

func (c *CO2) Set(b []byte) {
	*c = (CO2)(binary.BigEndian.Uint16(b))
}

// TVOC represents the current total volatile organic compounds value in ppb
type TVOC uint16

func (t TVOC) String() string {
	return strconv.Itoa(int(t)) + "ppb"
}

func (t *TVOC) Set(b []byte) {
	*t = (TVOC)(binary.BigEndian.Uint16(b))
}

// Env represents measurements from an environmental sensor.
type Env struct {
	CO2  CO2
	TVOC TVOC
}

// NewI2C returns an object that communicates over I2C to SGP30 environmental sensor.
//
// The address must be 0x58.
func NewI2C(b i2c.Bus, addr uint16) (*Dev, error) {
	switch addr {
	case I2CAddress:
	default:
		return nil, errors.New("sgp30: given address not supported by device")
	}
	d := &Dev{
		d: &i2c.Dev{Bus: b, Addr: addr},
		env: Env{
			CO2:  400,
			TVOC: 0,
		},
	}
	if err := d.makeDev(); err != nil {
		return nil, err
	}
	return d, nil
}

// Dev is a handle to an initialized SGP30 device.
type Dev struct {
	d   conn.Conn
	env Env
}

// EquivalentCO2 returns the latest equivalent CO2 value
func (d *Dev) EquivalentCO2() CO2 {
	return d.env.CO2
}

// TotalVOC returns the latest total VOC value
func (d *Dev) TotalVOC() TVOC {
	return d.env.TVOC
}

func (d *Dev) makeDev() error {
	// Sending  an "sgp30_iaq_init" command starts the air quality measurement
	err := d.initAirQuality()
	if err != nil {
		return err
	}

	// After the "sgp30_iaq_init" command, a "sgp30_measure_iaq" command has to be sent in regular
	// intervals of 1s to ensure proper operation of the dynamic baseline compensation algorithm.
	if err := d.measure(); err != nil {
		log.Print(err)
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			if err := d.measure(); err != nil {
				log.Print(err)
			}
		}
	}()

	return nil
}

func (d *Dev) initAirQuality() error {
	var err error
	if err = d.writeCommand(InitAirQuality); err == nil {
		time.Sleep(time.Second * 20)
	}

	return err
}

func (d *Dev) measure() error {
	buf := make([]byte, commandResponseLength[MeasureAirQuality])
	err := d.readCommand(MeasureAirQuality, buf)
	if err != nil {
		return err
	}
	d.env.CO2.Set(buf[0:2])
	d.env.TVOC.Set(buf[3:5])

	return nil
}

func (d *Dev) readCommand(cmd uint16, b []byte) error {
	if len(b) != commandResponseLength[cmd] {
		return errors.New("response length mismatch")
	}

	regAddr := []byte{byte(cmd >> 8), byte(cmd & 0xFF)}
	if err := d.d.Tx(regAddr, nil); err != nil {
		return err
	}
	time.Sleep(commandDuration[cmd])
	if err := d.d.Tx(nil, b); err != nil {
		return err
	}

	return nil
}

func (d *Dev) writeCommand(cmd uint16) error {
	regAddr := []byte{byte(cmd >> 8), byte(cmd & 0xFF)}
	if err := d.d.Tx(regAddr, nil); err != nil {
		return err
	}
	time.Sleep(commandDuration[cmd])
	return nil
}
