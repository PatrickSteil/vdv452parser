// Copyright 2015 geOps
// Authors: patrick.brosi@geops.de
//
// Use of this source code is governed by a GPL v2
// license that can be found in the LICENSE file

package vdv452

type Line struct {
	LineNo    int
	RouteAbbr string
	Direction int
	LineAbbr  string
	LineDesc  string
	RouteType int
	Sequence  []RouteSequence
}

type RouteSequence struct {
	SequenceNo  int
	PointType   int
	PointNo     int
	DestNo      int
	LineNode    bool
	Productive  int
	NoBoarding  bool
	NoAlighting bool
	RequestStop bool
}
