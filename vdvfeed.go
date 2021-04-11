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
	Stops map[uint64]*vdv452.Stop
}

// NewVDV452 creates a new, empty VDV452 feed
func NewVDV452() *VDV452 {
	g := VDV452{
		Stops: make(map[uint64]*vdv452.Stop),
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
		}
	}

	return e
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
		s.Longitude = float32(feed.getFloat("POINT_LONGITUDE", r, x10p.Cols))
		s.Latitude = float32(feed.getFloat("POINT_LATITUDE", r, x10p.Cols))
		fmt.Println(r)
		feed.Stops[uint64(s.Point_Type)*7000000+uint64(s.Point_No)] = s
	}
	return nil
}

func (feed *VDV452) getStr(name string, row []string, cols map[string]int) string {
	var idx int
	if _, ok := cols[name]; !ok {
		panic(fmt.Errorf("Missing column %s", name))
	}
	if idx >= len(row) {
		panic(fmt.Errorf("Missing column idx %d (%s)", idx, name))
	}
	return row[idx]
}

func (feed *VDV452) getFloat(name string, row []string, cols map[string]int) float64 {
	var idx int
	if _, ok := cols[name]; !ok {
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

func (feed *VDV452) getInt(name string, row []string, cols map[string]int) int {
	var idx int
	if _, ok := cols[name]; !ok {
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
