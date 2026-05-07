package store

import "gps-server/models"

var Devices = map[string]models.Device{
	// "354778340715856": {IMEI: "354778340715856", Name: "Jegan's Device"},
	"868377074976567": {IMEI: "868377074976567", Name: "4G Dashcam IMEI"},
	"377074976567":    {IMEI: "377074976567", Name: "4G Dashcam (JT808 ID)"},
}

func GetDevice(imei string) (models.Device, bool) {
	d, ok := Devices[imei]
	return d, ok
}
