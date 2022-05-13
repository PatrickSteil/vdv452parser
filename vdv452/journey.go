// Copyright 2015 geOps
// Authors: patrick.brosi@geops.de
//
// Use of this source code is governed by a GPL v2
// license that can be found in the LICENSE file

package vdv452

type Journey struct {
	JourneyNo     uint64
	DepartureTime int
	LineNo        int
	BlockNo       int
	DayTypeNo     int
	JourneyType   int
	TimingGroupNo int
	RouteAbbr     string
	TrainNo       int
}
