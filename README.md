# Airquality Exporter [![Integration](https://github.com/0x46616c6b/airquality-exporter/actions/workflows/integration.yml/badge.svg)](https://github.com/0x46616c6b/airquality-exporter/actions/workflows/integration.yml) [![Quality](https://github.com/0x46616c6b/airquality-exporter/actions/workflows/quality.yml/badge.svg)](https://github.com/0x46616c6b/airquality-exporter/actions/workflows/quality.yml)

The airquality-exporter provides metrics from environmental sensors to give insights about the airquality in a room. The exporter requires the environmental sensors [BME280](https://www.bosch-sensortec.com/products/environmental-sensors/humidity-sensors-bme280/) (for temperature, humidity and pressure) and [SGP30](https://www.sensirion.com/de/umweltsensoren/gassensoren/sgp30/) (carbon dioxide and volatile organic compounds).

## Development

Build the exporter using the Makefile

```shell
make build
```

Run the exporter

```shell
make run
```
