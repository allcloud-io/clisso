package yubikey

import (
	"github.com/karalabe/hid"
	log "github.com/sirupsen/logrus"
)

// IsAttached queries the connected USB devices and returns true if a YubiKey is attached
func IsAttached() bool {
	var yubiKeyVendorID uint16 = 0x1050

	// List all USB devices matching the YubiKey vendor ID
	devices := hid.Enumerate(yubiKeyVendorID, 0)

	if len(devices) == 0 {
		log.Debug("No YubiKey device detected")
		return false
	}

	// Log information about the detected YubiKey(s)
	if log.GetLevel() == log.DebugLevel {
		for _, device := range devices {
			log.WithFields(log.Fields{
				"vid":     device.VendorID,
				"pid":     device.ProductID,
				"product": device.Product,
			}).Debug("YubiKey device detected")
		}
	}
	return true
}
