/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

var testdata = []struct {
	app      int32
	provider int32
	result   int32
}{
	{0, 0, 3600},
	{7200, 0, 7200},
	{0, 7200, 7200},
	{7200, 14400, 7200},
}

func TestSessionDuration(t *testing.T) {
	for _, tc := range testdata {
		viper.Set("apps.test.duration", tc.app)
		viper.Set("providers.test.duration", tc.provider)

		res := sessionDuration("test", "test")
		if res != tc.result {
			t.Fatalf("Invalid duration: got %v, want: %v", res, tc.result)
		}
	}
}
