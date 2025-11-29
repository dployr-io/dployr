// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import "testing"

func TestAppErrorConstants(t *testing.T) {
	errors := map[AppError]string{
		BadRequest:   "bad_request",
		Unavailable:  "unavailable",
		RuntimeError: "runtime_error",
	}

	for err, expected := range errors {
		if string(err) != expected {
			t.Errorf("AppError %q != %q", err, expected)
		}
	}
}
