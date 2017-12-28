// Copyright 2016,2017 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package handler http server handler functions
package handler

import (
	"testing"
	"time"
)

var SEC_ERR = int64(10)

func TestParseSec(t *testing.T) {
	var err error
	var i int64

	nowSec := int64(time.Now().UTC().Unix())
	testcases := []struct {
		testStr string
		testSec int64
	}{
		{"-2000ms", nowSec - 2},
		{"-60s", nowSec - 60},
		{"-3mn", nowSec - 3*60},
		{"-8h", nowSec - 8*60*60},
		{"-1d", nowSec - 24*60*60},
		{"2000ms", 2},
		{"60s", 60},
		{"3mn", 3 * 60},
		{"8h", 8 * 60 * 60},
		{"1d", 24 * 60 * 60},
		{"42000", 42},
	}

	// run test list
	for _, tc := range testcases {
		if i, err = parseSec(tc.testStr); err != nil {
			t.Errorf("error parsing '%s'", tc.testStr)
		}
		if i < (tc.testSec-SEC_ERR) || i > (tc.testSec+SEC_ERR) {
			t.Errorf("error parsing '%s' [%d != %d]", tc.testStr, i, tc.testSec)
		}
	}

	// check that we raise errors on bad strings
	if i, err = parseSec(""); err == nil {
		t.Errorf("no error while parsing empty string \"\"")
	}
}
