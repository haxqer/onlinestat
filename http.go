package main

import (
	"encoding/json"
	"net/http"
)

func ServeOnlineUserList(w http.ResponseWriter, r *http.Request) {
	onlineIndex := StatData.GetOnlineIndex()
	sList := map[string]*UserStat{}
	for k := range onlineIndex {
		v := StatData.Get(k)
		if v == nil {
			continue
		}
		sList[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sList)
}

func ServeTotalUserList(w http.ResponseWriter, r *http.Request) {
	totalIndex := StatData.GetTotalIndex()
	sList := map[string]*UserStat{}
	for k := range totalIndex {
		v := StatData.Get(k)
		if v == nil {
			continue
		}
		sList[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sList)
}