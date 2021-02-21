package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"

	"github.com/0x46616c6b/airquality-exporter/devices/sgp30"
)

var (
	addr = flag.String("listen-address", ":9229", "The address to listen on for HTTP requests.")

	temperature = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "airquality",
		Name:      "temperature",
		Help:      "Current temperature (°C)",
	})
	humidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "airquality",
		Name:      "humidity",
		Help:      "Current humidity (%)",
	})
	pressure = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "airquality",
		Name:      "pressure",
		Help:      "Current pressure (hPa)",
	})
	co2 = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "airquality",
		Name:      "co2",
		Help:      "Amount of CO2 in air (ppm)",
	})
	tvoc = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "airquality",
		Name:      "tvoc",
		Help:      "Amount of total volatile organic compound in air (ppb)",
	})
)

func init() {
	prometheus.MustRegister(temperature, humidity, pressure, co2, tvoc)
	flag.Parse()
}

func main() {
	log.Println("Initializing sensors")
	// Load all the drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a handle to the first available I2C bus
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = bus.Close()
	}()

	// Open a handle to a bme280 connected on the I2C bus
	bme280, err := bmxx80.NewI2C(bus, 0x76, &bmxx80.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = bme280.Halt()
	}()

	// Open a handle to a sgp30 connected on the I2C bus
	sgp30, err := sgp30.NewI2C(bus, sgp30.I2CAddress)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Sensors initialized")

	tick := time.Tick(time.Second * 10)
	go func() {
		for {
			<-tick
			// Read temperature from the sensor.
			var env physic.Env
			if err = bme280.Sense(&env); err != nil {
				log.Fatal(err)
			}

			// Values from BME280 environmental sensor.
			temperature.Set(env.Temperature.Celsius())
			// Pressure is represented as Nano Pascal. Needs to convert as Hecto Pascal.
			p := float64(env.Pressure) / float64(physic.Pascal) / 100
			pressure.Set(p)
			// Humidity is represented as Milli Percent.
			h := float64(env.Humidity) / float64(physic.PercentRH)
			humidity.Set(h)

			co2.Set(float64(sgp30.EquivalentCO2()))
			tvoc.Set(float64(sgp30.TotalVOC()))
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
