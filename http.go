package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type UserOutputStat struct {
	TotalOnline int64
	TodayOnline int64
}

func isBeforeDawn(gap int64) bool {
	if gap >=0 && gap < calInterval * 3 {
		return true
	}
	return false
}

func ServeOnlineUserList(w http.ResponseWriter, r *http.Request) {
	todayBeginning := GetTodayBeginning()
	nowTs := time.Now().Unix()
	gap := nowTs-todayBeginning
	
	onlineIndex := StatData.GetOnlineIndex()
	sList := map[string]*UserOutputStat{}
	for k := range onlineIndex {
		v := StatData.Get(k)
		if v == nil {
			continue
		}

		todayOnline := v.TodayOnline
		if isBeforeDawn(gap) {
			todayOnline = 0
		}
		sList[k] = &UserOutputStat{
			TotalOnline: v.TotalOnline,
			TodayOnline: todayOnline,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sList)
}

func ServeTotalUserList(w http.ResponseWriter, r *http.Request) {
	todayBeginning := GetTodayBeginning()
	nowTs := time.Now().Unix()
	gap := nowTs-todayBeginning

	totalIndex := StatData.GetTotalIndex()
	sList := map[string]*UserOutputStat{}
	for k := range totalIndex {
		v := StatData.Get(k)
		if v == nil {
			continue
		}
		todayOnline := v.TodayOnline
		if isBeforeDawn(gap) {
			todayOnline = 0
		}
		sList[k] = &UserOutputStat{
			TotalOnline: v.TotalOnline,
			TodayOnline: todayOnline,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sList)
}