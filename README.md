The kinX blaster tool reads/writes the EEPROM of the kinX’s CY7C65632 USB 2.0
hub.

## Usage

```
mk66f-blaster        # read the EEPROM
mk66f-blaster -raw   # print the raw EEPROM bytes to stdout instead of parsing
mk66f-blaster -write # write the default config to the EEPROM
```

## Installation

```
go get -u github.com/kinx-project/mk66f-blaster
```
