// Copyright 2015 geOps
// Authors: patrick.brosi@geops.de
//
// Use of this source code is governed by a GPL v2
// license that can be found in the LICENSE file

package vdv452

// Description of locations. All the points on the network are contained in this
// relation. There is also a description of how the network points are formed
// into area groups. A bus stop / depot can be made up of several stopping
// points (e.g. when travelling a route in both directions). In this relation, this is
// highlighted by references between the network points which belong
// together. A bus stop / depot can have a maximum of 100 stopping points
// assigned to it. No stopping points with the same number are allowed for one
// bus stop/depot. The code (STOP_ABBR (ORT_REF_ORT_KUERZEL)) and
// the (STOP_NO (ORT_REF_ORT)) number must be unique across all stops
// and depots.
type Stop struct {
	Point_Type            int8
	Point_No              uint
	Point_Desc            string
	Stop_No               uint
	Stop_Point_No               uint
	Stop_Type             int8
	Stop_Long_No          int
	Stop_Abbr             string
	Stop_Desc             string
	Stop_Point_Desc             string
	Zone_Cell_No          int
	Longitude             float32
	Latitude              float32
	Elevation             int
	Stop_No_Local         int
	Stop_No_National      int
	Stop_No_International int
}
