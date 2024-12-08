// Copyright 2015 geOps
// Authors: patrick.brosi@geops.de
//
// Use of this source code is governed by a GPL v2
// license that can be found in the LICENSE file

package vdv452

type Void struct{}

type DayType struct {
	DayTypeNo     int
	OperatingDays map[uint64]Void
	DayTypeDesc string
}
