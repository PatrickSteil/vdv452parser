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
	"sort"
	"strconv"
	"strings"
)

var TRANS = map[string]string{
	"POINT_TYPE":      "ONR_TYP_NR",
	"POINT_NO":        "ORT_NR",
	"POINT_DESC":      "ORT_NAME",
	"STOP_NO":         "ORT_REF_ORT",
	"STOP_TYPE":       "ORT_REF_ORT_TYP",
	"STOP_ABBR":       "ORT_REF_ORT_KUERZEL",
	"STOP_DESC":       "ORT_REF_ORT_NAME",
	"POINT_LATITUDE":  "ORT_POS_LAENGE",
	"POINT_LONGITUDE": "ORT_POS_HOEHE",

	"JOURNEY_NO":      "FRT_FID",
	"DEPARTURE_TIME":  "FRT_START",
	"LINE_NO":         "LI_NR",
	"ROUTE_ABBR":      "STR_LI_VAR",
	"DAY_TYPE_NO":     "TAGESART_NR",
	"JOURNEY_TYPE_NO": "FAHRTART_NR",
	"TIMING_GROUP_NO": "FGR_NR",
	"BLOCK_NO":        "UM_UID",

	"OPERATING_DAY": "BETRIEBSTAG",

	"SEQUENCE_NO":  "LI_LFD_NR",
	"DEST_NO":      "ZNR_NR",
	"LINE_NODE":    "LI_KNOTEN",
	"PRODUCTIVE":   "PRODUKTIV",
	"NO_BOARDING":  "EINSTEIGEVERBOT",
	"NO_ALIGHTING": "AUSSTEIGEVERBOT",
	"REQUEST_STOP": "BEDARFSHALT",

	"DIRECTION":  "LI_RI_NR",
	"LINE_ABBR":  "LI_KUERZEL",
	"OP_DEP_NO":  "BEREICH_NR",
	"LINE_DESC":  "LIDNAME",
	"ROUTE_TYPE": "ROUTEN_ART",

	"OP_DEP_ABBR":        "STR_BEREICH",
	"OP_DEP_DESC":        "BEREICH_TEXT",
	"COMPANY":            "UNTERNEHMEN",
	"COMPANY_ABBR":       "ABK_UNTERNEHMEN",
	"BUSINESS_AREA_DESC": "BETRIEBSGEBIET_BEZ",
	"VH_TYPE_NO":         "FZG_TYP_NR",
	"VEHICLE_NO":         "FZG_NR",
	"VEHICLE_TYPE":       "FZG_TYP_NR",

	"VH_TYPE_DESC":      "FZG_TYP_TEXT",
	"VH_TYPE_ABBR":      "STR_FZG_TYP",
	"VH_TYPE_SPEC_SEAT": "FZG_TYP_SITZ",

	"DEST_BRIEF_TEXT": "FAHRERKURZTEXT",
	"DEST_SIDE_TEXT":  "SEITENTEXT",
	"DEST_FRONT_TEXT": "ZNR_TEXT",

	"TO_POINT_TYPE":   "SEL_ZIEL_TYP",
	"TO_POINT_NO":     "SEL_ZIEL",
	"FROM_POINT_TYPE": "ONR_TYPE_NR",
	"FROM_POINT_NO":   "ORT_NR",
	"LINK_DISTANCE":   "SEL_LAENGE",
	"TRAVEL_TIME":     "SEL_FZT",
}

type VDV452 struct {
	Stops                map[uint64]*vdv452.Stop
	Lines                map[string]*vdv452.Line
	TravelTimes          map[uint64]map[uint64]int
	WaitTimes            map[uint64]map[uint64]int
	Journeys             map[uint64]*vdv452.Journey
	DayTypes             map[uint64]*vdv452.DayType
	Destinations         map[uint64]*vdv452.Destination
	VehicleTypes         map[uint64]*vdv452.VehicleType
	Vehicles             map[uint64]*vdv452.Vehicle
	Companies            map[uint64]*vdv452.Company
	OperatingDepartments map[uint64]*vdv452.OperatingDepartment
	Blocks               map[uint64]*vdv452.Block
}

type seqAsc []vdv452.RouteSequence

func (v seqAsc) Len() int           { return len(v) }
func (v seqAsc) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v seqAsc) Less(i, j int) bool { return v[i].SequenceNo < v[j].SequenceNo }

// NewVDV452 creates a new, empty VDV452 feed
func NewVDV452() *VDV452 {
	g := VDV452{
		Stops:                make(map[uint64]*vdv452.Stop),
		Lines:                make(map[string]*vdv452.Line),
		TravelTimes:          make(map[uint64]map[uint64]int),
		WaitTimes:            make(map[uint64]map[uint64]int),
		Journeys:             make(map[uint64]*vdv452.Journey),
		DayTypes:             make(map[uint64]*vdv452.DayType),
		Destinations:         make(map[uint64]*vdv452.Destination),
		VehicleTypes:         make(map[uint64]*vdv452.VehicleType),
		Vehicles:             make(map[uint64]*vdv452.Vehicle),
		Companies:            make(map[uint64]*vdv452.Company),
		OperatingDepartments: make(map[uint64]*vdv452.OperatingDepartment),
		Blocks:               make(map[uint64]*vdv452.Block),
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

			if x10p.TblName == "REC_ORT" {
				feed.parseStop(&x10p)
			}

			if x10p.TblName == "ROUTE_SEQUENCE" {
				feed.parseRouteSequence(&x10p)
			}

			if x10p.TblName == "LID_VERLAUF" {
				feed.parseRouteSequence(&x10p)
			}

			if x10p.TblName == "LINE" {
				feed.parseLine(&x10p)
			}

			if x10p.TblName == "REC_LID" {
				feed.parseLine(&x10p)
			}

			if x10p.TblName == "TRAVEL_TIME" {
				feed.parseTravelTime(&x10p)
			}

			if x10p.TblName == "SEL_FZT_FELD" {
				feed.parseTravelTime(&x10p)
			}

			if x10p.TblName == "WAIT_TIME" {
				feed.parseWaitTime(&x10p)
			}

			if x10p.TblName == "ORT_HZTF" {
				feed.parseWaitTime(&x10p)
			}

			if x10p.TblName == "JOURNEY" {
				feed.parseJourney(&x10p)
			}

			if x10p.TblName == "REC_FRT" {
				feed.parseJourney(&x10p)
			}

			if x10p.TblName == "PERIOD" {
				feed.parsePeriod(&x10p)
			}

			if x10p.TblName == "FIRMENKALENDER" {
				feed.parsePeriod(&x10p)
			}

			if x10p.TblName == "DESTINATION" {
				feed.parseDestination(&x10p)
			}

			if x10p.TblName == "REC_ZNR" {
				feed.parseDestination(&x10p)
			}

			if x10p.TblName == "VEHICLE_TYPE" {
				feed.parseVehicleType(&x10p)
			}

			if x10p.TblName == "MENGE_FZG_TYP" {
				feed.parseVehicleType(&x10p)
			}

			if x10p.TblName == "VEHICLE" {
				feed.parseVehicle(&x10p)
			}

			if x10p.TblName == "FAHRZEUG" {
				feed.parseVehicle(&x10p)
			}

			if x10p.TblName == "TRANSPORT_COMPANY" {
				feed.parseCompany(&x10p)
			}

			if x10p.TblName == "ZUL_VERKEHRSBETRIEB" {
				feed.parseCompany(&x10p)
			}

			if x10p.TblName == "BLOCK" {
				feed.parseBlock(&x10p)
			}

			if x10p.TblName == "REC_UMLAUF" {
				feed.parseBlock(&x10p)
			}

			if x10p.TblName == "OPERATING_DEPARTMENT" {
				feed.parseOpDep(&x10p)
			}

			if x10p.TblName == "MENGE_BEREICH" {
				feed.parseOpDep(&x10p)
			}
		}
	}

	return e
}

func (feed *VDV452) parseLine(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		lineNo := feed.getInt("LINE_NO", r, x10p.Cols)
		routeAbbr := feed.getStr("ROUTE_ABBR", r, x10p.Cols)

		lineId := fmt.Sprintf("%06d", lineNo) + "." + routeAbbr
		fmt.Println("Parsed line", lineId)
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

		lineId := fmt.Sprintf("%06d", lineNo) + "." + routeAbbr
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

	for id, l := range feed.Lines {
		fmt.Println(id, len(l.Sequence))
		sort.Sort(seqAsc(l.Sequence))
	}
	return nil
}

func (feed *VDV452) parseWaitTime(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		tGroupNo := uint64(feed.getInt("TIMING_GROUP_NO", r, x10p.Cols))
		pointType := uint64(feed.getInt("POINT_TYPE", r, x10p.Cols))
		pointNo := uint64(feed.getInt("POINT_NO", r, x10p.Cols))
		t := feed.getInt("WAIT_TIME", r, x10p.Cols)

		tGroup := tGroupNo
		ft := pointType*1000000 + pointNo

		if _, ok := feed.WaitTimes[tGroup]; !ok {
			feed.WaitTimes[tGroup] = make(map[uint64]int, 0)
		}
		feed.WaitTimes[tGroup][ft] = t
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
		j.BlockNo = (feed.getInt("BLOCK_NO", r, x10p.Cols))
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
		v.VhTypeSpecSeat = (feed.getInt("VH_TYPE_SPEC_SEAT", r, x10p.Cols))
		feed.VehicleTypes[v.VhTypeNo] = v
	}
	return nil
}

func (feed *VDV452) parseOpDep(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		v := new(vdv452.OperatingDepartment)
		v.OpDepNo = (feed.getInt("OP_DEP_NO", r, x10p.Cols))
		v.OpDepAbbr = (feed.getStr("OP_DEP_ABBR", r, x10p.Cols))
		v.OpDepDesc = (feed.getStr("OP_DEP_DESC", r, x10p.Cols))
		feed.OperatingDepartments[uint64(v.OpDepNo)] = v
	}
	return nil
}

func (feed *VDV452) parseCompany(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		v := new(vdv452.Company)
		v.CompanyNo = (feed.getInt("COMPANY", r, x10p.Cols))
		v.CompanyAbbr = (feed.getStr("COMPANY_ABBR", r, x10p.Cols))
		v.BusinessAreaDesc = (feed.getStr("BUSINESS_AREA_DESC", r, x10p.Cols))
		feed.Companies[uint64(v.CompanyNo)] = v
	}
	return nil
}

func (feed *VDV452) parseBlock(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		bl := new(vdv452.Block)
		bl.DayTypeNo = (feed.getInt("DAY_TYPE_NO", r, x10p.Cols))
		bl.BlockNo = (feed.getInt("BLOCK_NO", r, x10p.Cols))
		bl.VhTypeNo = (feed.getInt("VH_TYPE_NO", r, x10p.Cols))

		id := bl.DayTypeNo*1000 + bl.BlockNo

		feed.Blocks[uint64(id)] = bl
	}
	return nil
}

func (feed *VDV452) parseVehicle(x10p *x10parser.X10Parser) (err error) {
	for r, _ := x10p.Row(); len(r) > 0; r, _ = x10p.Row() {
		v := new(vdv452.Vehicle)
		v.VehicleNo = (feed.getInt("VEHICLE_NO", r, x10p.Cols))
		v.VehicleType = (feed.getInt("VEHICLE_TYPE", r, x10p.Cols))
		v.Company = (feed.getInt("COMPANY", r, x10p.Cols))
		feed.Vehicles[uint64(v.VehicleNo)] = v
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

		lat := feed.getInt("POINT_LATITUDE", r, x10p.Cols)
		lon := feed.getInt("POINT_LONGITUDE", r, x10p.Cols)

		s.Longitude = float32(lon/10000000) + float32((lon%10000000/100000))/60.0 + (float32(lon%10000000%100000)/1000.0)/3600.0

		s.Latitude = float32(lat/10000000) + float32((lat%10000000/100000))/60.0 + (float32(lat%10000000%100000)/1000.0)/3600.0

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
			dt.DayTypeNo = int(dayTypeId)
			feed.DayTypes[dayTypeId] = dt
		} else {
			dt = dayType
		}
		dt.OperatingDays[opDay] = struct{}{}
	}
	return nil
}

func (feed *VDV452) getStr(name string, row []string, cols map[string]int) string {
	var ok bool
	if _, ok = cols[name]; !ok {
		if trans, ok := TRANS[name]; ok {
			// try german translation
			name = trans
		}
	}

	var idx int
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
	if _, ok = cols[name]; !ok {
		if trans, ok := TRANS[name]; ok {
			// try german translation
			name = trans
		}
	}
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
	if _, ok = cols[name]; !ok {
		if trans, ok := TRANS[name]; ok {
			// try german translation
			name = trans
		}
	}
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
	if _, ok = cols[name]; !ok {
		if trans, ok := TRANS[name]; ok {
			// try german translation
			name = trans
		}
	}
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
