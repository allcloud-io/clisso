/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package onelogin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDevice(t *testing.T) {

	var deviceList = []Device{
		{
			DeviceID:   01,
			DeviceType: "Yubico YubiKey",
		},
		{
			DeviceID:   02,
			DeviceType: "OneLogin Protect",
		},
		{
			DeviceID:   03,
			DeviceType: "Google Authenticator",
		},
	}

	cases := []struct {
		Name           string
		Devices        []Device
		Opts           *DeviceOptions
		ExpectedDevice *Device
		ExpectedError  error
	}{
		{
			Name:           "NoDevices",
			Devices:        []Device{},
			Opts:           &DeviceOptions{},
			ExpectedDevice: nil,
			ExpectedError:  errors.New("no MFA device returned by Onelogin"),
		},
		{
			Name:           "AutodetectYubiKey",
			Devices:        deviceList,
			Opts:           &DeviceOptions{IsYubiKeyAutoDetected: true},
			ExpectedDevice: &Device{DeviceID: 01, DeviceType: "Yubico YubiKey"},
			ExpectedError:  nil,
		},
		{
			Name:           "SelectedMfaDevice",
			Devices:        deviceList,
			Opts:           &DeviceOptions{MfaDevice: "Google Authenticator"},
			ExpectedDevice: &Device{DeviceID: 03, DeviceType: "Google Authenticator"},
			ExpectedError:  nil,
		},
		{
			Name:           "SelectedMfaDeviceOverride",
			Devices:        deviceList,
			Opts:           &DeviceOptions{IsYubiKeyAutoDetected: true, MfaDevice: "Google Authenticator"},
			ExpectedDevice: &Device{DeviceID: 03, DeviceType: "Google Authenticator"},
			ExpectedError:  nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {

			d, err := getDevice(c.Devices, c.Opts)
			assert.Equal(t, c.ExpectedDevice, d)
			assert.Equal(t, c.ExpectedError, err)
		})
	}
}
