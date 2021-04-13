// Copyright 2016 Patrick Brosi
// Authors: info@patrickbrosi.de
//
// Use of this source code is governed by a GPL v2
// license that can be found in the LICENSE file

package vdv452parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"patrickbrosi.de/vdv452parser/vdv452"
	"patrickbrosi.de/x10parser"
	"strconv"
	"strings"
)

type VDV452 struct {
	Stops        map[uint64]*vdv452.Stop
	Lines        map[string]*vdv452.Line
	TravelTimes  map[uint64]map[uint64]int
	Journeys     map[uint64]*vdv452.Journey
	DayTypes     map[uint64]*vdv452.DayType
	Destinations map[uint64]*vdv452.Destination
	VehicleTypes map[uint64]*vdv452.VehicleType
}

// NewVDV452 creates a new, empty VDV452 feed
func NewVDV452() *VDV452 {
	g := VDV452{
		Stops:        make(map[uint64]*vdv452.Stop),
		Lines:        make(map[string]*vdv452.Line),
		TravelTimes:  make(map[uint64]map[uint64]int),
		Journeys:     make(map[uint64]*vdv452.Journey),
		DayTypes:     make(map[uint64]*vdv452.DayType),
		Destinations: make(map[uint64]*vdv452.Destination),
		VehicleTypes: make(map[uint64]*vdv452.VehicleType),
	}
	return &g
}

func (feed *VDV452) Parse(path string) error {
	var e error
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if !item.IsDir() {
			// handle file there
			fullpath := filepath.Join(path, item.Name())
			x10p := x10parser.X10Parser{}
			e = x10p.Open(fullpath)

			if x10p.TblName == "STOP" {
				feed.parseStop(&x10p)
			}

			if x10p.TblName == "ROUTE_SEQUENCE" {
				feed.parseRouteSequence(&x10p)
			}

			if x10p.TblName == "LINE" {
				feed.parseLine(&x10p)
			}

			if x10p.TblName == "TRAVEL_TIME" {
				feed.parseTravelTime(&x10p)
			}

			if x10p.TblName == "JOURNEY" {
				feed.parseJourney(&x10p)
			}

			if x10p.TblName == "PERIOD" {
				feed.parsePeriod(&x10p)
			}

			if x10p.TblName == "DESTINATION" {
				feed.parseDestination(&x10p)
			}

			if x10p.TblName == "VEHICLE_TYPE" {
				feed.parseVehicleType(&x10p)
			}
		}
	}

	return e
}

func (feed *VDV452) parseLine(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		lineNo := feed.getInt("LINE_NO", r, x10p.Cols)
		routeAbbr := feed.getStr("ROUTE_ABBR", r, x10p.Cols)

		lineId := fmt.Sprintf("%06d", lineNo) + routeAbbr
		var l *vdv452.Line
		if line, ok := feed.Lines[lineId]; !ok {
			l = new(vdv452.Line)
			feed.Lines[lineId] = l
		} else {
			l = line
		}
		l.LineNo = lineNo
		l.RouteAbbr = feed.getStr("ROUTE_ABBR", r, x10p.Cols)
		l.Direction = int(feed.getInt("DIRECTION", r, x10p.Cols))
		l.LineAbbr = feed.getStr("LINE_ABBR", r, x10p.Cols)
		l.OpDepNo = feed.getInt("OP_DEP_NO", r, x10p.Cols)
		l.LineDesc = feed.getStr("LINE_DESC", r, x10p.Cols)
		l.RouteType = int(feed.getInt("ROUTE_TYPE", r, x10p.Cols))
	}
	return nil
}

func (feed *VDV452) parseRouteSequence(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		lineNo := feed.getInt("LINE_NO", r, x10p.Cols)
		routeAbbr := feed.getStr("ROUTE_ABBR", r, x10p.Cols)

		lineId := fmt.Sprintf("%06d", lineNo) + routeAbbr
		var l *vdv452.Line
		if line, ok := feed.Lines[lineId]; !ok {
			l = new(vdv452.Line)
			feed.Lines[lineId] = l
		} else {
			l = line
		}
		rs := vdv452.RouteSequence{}

		rs.SequenceNo = int(feed.getInt("SEQUENCE_NO", r, x10p.Cols))
		rs.PointType = int(feed.getInt("POINT_TYPE", r, x10p.Cols))
		rs.PointNo = int(feed.getInt("POINT_NO", r, x10p.Cols))
		rs.DestNo = int(feed.getInt("DEST_NO", r, x10p.Cols))
		rs.LineNode = feed.getBool("LINE_NODE", r, x10p.Cols, true, false)
		// rs.Productive = feed.getBool("PRODUCTIVE", r, x10p.Cols, )
		rs.NoBoarding = feed.getBool("NO_BOARDING", r, x10p.Cols, false, false)
		rs.NoAlighting = feed.getBool("NO_ALIGHTING", r, x10p.Cols, false, false)
		rs.RequestStop = feed.getBool("REQUEST_STOP", r, x10p.Cols, false, false)

		l.Sequence = append(l.Sequence, rs)
	}
	return nil
}

func (feed *VDV452) parseTravelTime(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		opDepNo := uint64(feed.getInt("OP_DEP_NO", r, x10p.Cols))
		tGroupNo := uint64(feed.getInt("TIMING_GROUP_NO", r, x10p.Cols))
		fromPointType := uint64(feed.getInt("POINT_TYPE", r, x10p.Cols))
		fromPointNo := uint64(feed.getInt("POINT_NO", r, x10p.Cols))
		toPointType := uint64(feed.getInt("TO_POINT_TYPE", r, x10p.Cols))
		toPointNo := uint64(feed.getInt("TO_POINT_NO", r, x10p.Cols))
		t := feed.getInt("TRAVEL_TIME", r, x10p.Cols)

		tGroup := opDepNo*1000000000 + tGroupNo
		ft := fromPointType*100000000000000 + fromPointNo*100000000 + toPointType*1000000 + toPointNo

		if _, ok := feed.TravelTimes[tGroup]; !ok {
			feed.TravelTimes[tGroup] = make(map[uint64]int, 0)
		}
		feed.TravelTimes[tGroup][ft] = t
	}
	return nil
}

func (feed *VDV452) parseJourney(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		j := new(vdv452.Journey)
		j.JourneyNo = uint64(feed.getInt("JOURNEY_NO", r, x10p.Cols))
		j.DepartureTime = (feed.getInt("DEPARTURE_TIME", r, x10p.Cols))
		j.LineNo = (feed.getInt("LINE_NO", r, x10p.Cols))
		j.RouteAbbr = (feed.getStr("ROUTE_ABBR", r, x10p.Cols))
		j.DayTypeNo = (feed.getInt("DAY_TYPE_NO", r, x10p.Cols))
		j.JourneyType = (feed.getInt("JOURNEY_TYPE_NO", r, x10p.Cols))
		j.TimingGroupNo = (feed.getInt("TIMING_GROUP_NO", r, x10p.Cols))
		// j.TrainNo = (feed.getInt("TRAIN_NO", r, x10p.Cols))

		feed.Journeys[j.JourneyNo] = j
	}
	return nil
}

func (feed *VDV452) parseDestination(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		d := new(vdv452.Destination)
		d.DestNo = uint64(feed.getInt("DEST_NO", r, x10p.Cols))
		d.DestBriefText = (feed.getStr("DEST_BRIEF_TEXT", r, x10p.Cols))
		d.DestSideText = (feed.getStr("DEST_SIDE_TEXT", r, x10p.Cols))
		d.DestFrontText = (feed.getStr("DEST_FRONT_TEXT", r, x10p.Cols))
		feed.Destinations[d.DestNo] = d
	}
	return nil
}

func (feed *VDV452) parseVehicleType(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		v := new(vdv452.VehicleType)
		v.VhTypeNo = uint64(feed.getInt("VH_TYPE_NO", r, x10p.Cols))
		v.VhTypeDesc = (feed.getStr("VH_TYPE_DESC", r, x10p.Cols))
		v.VhTypeAbbr = (feed.getStr("VH_TYPE_ABBR", r, x10p.Cols))
		feed.VehicleTypes[v.VhTypeNo] = v
	}
	return nil
}

func (feed *VDV452) parseStop(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		s := new(vdv452.Stop)
		s.Point_Type = int8(feed.getInt("POINT_TYPE", r, x10p.Cols))
		s.Point_No = uint(feed.getInt("POINT_NO", r, x10p.Cols))
		s.Point_Desc = feed.getStr("POINT_DESC", r, x10p.Cols)
		s.Stop_No = uint(feed.getInt("STOP_NO", r, x10p.Cols))
		s.Stop_Type = int8(feed.getInt("STOP_TYPE", r, x10p.Cols))
		s.Stop_Abbr = feed.getStr("STOP_ABBR", r, x10p.Cols)
		s.Stop_Desc = feed.getStr("STOP_DESC", r, x10p.Cols)
		s.Longitude = float32(feed.getFloat("POINT_LONGITUDE", r, x10p.Cols) / 10000000.0)
		s.Latitude = float32(feed.getFloat("POINT_LATITUDE", r, x10p.Cols) / 10000000.0)

		feed.Stops[uint64(s.Point_Type)*7000000+uint64(s.Point_No)] = s
	}
	return nil
}

func (feed *VDV452) parsePeriod(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		var dt *vdv452.DayType
		dayTypeId := uint64(feed.getInt("DAY_TYPE_NO", r, x10p.Cols))
		opDay := uint64(feed.getInt("OPERATING_DAY", r, x10p.Cols))
		if dayType, ok := feed.DayTypes[dayTypeId]; !ok {
			dt = new(vdv452.DayType)
			dt.OperatingDays = make(map[uint64]vdv452.Void)
			feed.DayTypes[dayTypeId] = dt
		} else {
			dt = dayType
		}
		dt.OperatingDays[opDay] = struct{}{}
	}
	return nil
}

func (feed *VDV452) getStr(name string, row []string, cols map[string]int) string {
	var idx int
	var ok bool
	if idx, ok = cols[name]; !ok {
		panic(fmt.Errorf("Missing column %s", name))
	}
	if idx >= len(row) {
		panic(fmt.Errorf("Missing column idx %d (%s)", idx, name))
	}
	return row[idx]
}

func (feed *VDV452) getFloat(name string, row []string, cols map[string]int) float64 {
	var idx int
	var ok bool
	if idx, ok = cols[name]; !ok {
		panic(fmt.Errorf("Missing column %s", name))
	}
	if idx >= len(row) {
		panic(fmt.Errorf("Missing column idx %d (%s)", idx, name))
	}
	if val := row[idx]; len(strings.TrimSpace(val)) > 0 {
		trimmed := strings.TrimSpace(val)
		num, err := strconv.ParseFloat(trimmed, 32)
		if err != nil {
			panic(fmt.Errorf("Expected float for field '%s', found '%s'", name, val))
		}
		return num
	} else {
		panic(fmt.Errorf("Expected required field '%s'", name))
	}
}

func (feed *VDV452) getBool(name string, row []string, cols map[string]int, req bool, def bool) bool {
	var idx int
	var ok bool
	if idx, ok = cols[name]; !ok {
		if !req {
			return def
		}
		panic(fmt.Errorf("Missing column %s", name))
	}
	if idx >= len(row) {
		if !req {
			return def
		}
		panic(fmt.Errorf("Missing column idx %d (%s)", idx, name))
	}
	if val := row[idx]; len(strings.TrimSpace(val)) > 0 {
		num, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			panic(fmt.Errorf("Expected bool for column '%s', found '%s'", name, val))
		}
		if num < 0 || num > 1 {
			panic(fmt.Errorf("Expected bool for column '%s', found '%s'", name, val))
		}
		if num == 1 {
			return true
		} else {
			return false
		}
	} else {
		if !req {
			return def
		}
		panic(fmt.Errorf("Expected required field '%s'", name))
	}
}

func (feed *VDV452) getInt(name string, row []string, cols map[string]int) int {
	var idx int
	var ok bool
	if idx, ok = cols[name]; !ok {
		panic(fmt.Errorf("Missing column %s", name))
	}
	if idx >= len(row) {
		panic(fmt.Errorf("Missing column idx %d (%s)", idx, name))
	}
	if val := row[idx]; len(strings.TrimSpace(val)) > 0 {
		num, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			panic(fmt.Errorf("Expected integer for column '%s', found '%s'", name, val))
		}
		return num
	} else {
		panic(fmt.Errorf("Expected required field '%s'", name))
	}
}
