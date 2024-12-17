package yubikey

import (
	"github.com/allcloud-io/clisso/log"
	"github.com/karalabe/hid"
)

// USB Vendor ID is a permanent ID issued by USB Implementers Forum
const yubiKeyVendorID uint16 = 0x1050

// IsAttached queries the connected USB devices and returns true if a YubiKey is attached
func IsAttached() bool {

	// List all USB devices matching the YubiKey vendor ID
	devices, err := hid.Enumerate(yubiKeyVendorID, 0)
	if err != nil {
		log.WithError(err).Error("Failed to enumerate USB devices")
		return false
	}

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
