package store

import "gps-server/models"

var Devices = map[string]models.Device{
	"359339075496001": {IMEI: "359339075496001", Name: "Test Device"},
	"354778340715856": {IMEI: "354778340715856", Name: "Jegan's Device"},
}

func GetDevice(imei string) (models.Device, bool) {
	d, ok := Devices[imei]
	return d, ok
}
