package main

import (
	"fmt"
	"github.com/haxqer/gintools/file"
	"sync"
	"time"
)

type StatDataStruct struct {
	mu             sync.RWMutex
	Data           map[string]*UserStat
	OnlineIndex    map[string]int
	ApiOnlineIndex map[string]int
}

type UserStat struct {
	OnlineAt    int64
	OfflineAt   int64
	CalculateAt int64
	TotalOnline int64
	TodayOnline int64
}

const (
	location      = "Asia/Shanghai"
	fileName      = "data/a.store" // 持久化文件的路径, 可以为 相对路径或者绝对路径
	calInterval   = 1              // 计算行为发生的间隔, 单位为 秒
	storeInterval = 30             // 持久化行为发生的时间间隔, 单位为 秒
	clearGap      = 86400
)

var (
	shanghai     *time.Location
	StatData     *StatDataStruct
	calLimiter   <-chan time.Time
	storeLimiter <-chan time.Time
	clearLimiter <-chan time.Time
)

func init() {
	var err error
	shanghai, err = time.LoadLocation(location)
	if err != nil {
		panic(err)
	}

	StatData = &StatDataStruct{
		Data:        make(map[string]*UserStat),
		OnlineIndex: make(map[string]int),
	}

	if notExist := file.CheckNotExist(fileName); !notExist {
		if err := load(fileName, StatData); err != nil {
			panic(err)
		}
	} else {
		err = store(fileName, StatData.Snapshot())
		if err != nil {
			panic(err)
		}
	}

	tickClearData()
	tickCalData()
	tickStoreData()
}

func tickClearData() {
	StatData.ClearData()
	interval := 3 * time.Hour
	clearLimiter = time.After(interval)
	go func() {
		for {
			<-clearLimiter
			StatData.ClearData()
			clearLimiter = time.After(interval)
		}
	}()
}

func tickCalData() {
	interval := calInterval * time.Second
	calLimiter = time.After(interval)
	go func() {
		for {
			<-calLimiter
			StatData.Ticker()
			calLimiter = time.After(interval)
		}
	}()
}

func tickStoreData() {
	interval := storeInterval * time.Second
	storeLimiter = time.After(interval)
	go func() {
		for {
			<-storeLimiter
			err := store(fileName, StatData.Snapshot())
			if err != nil {
				fmt.Printf("%+v", err)
			}
			storeLimiter = time.After(interval)
		}
	}()
}

func (s *StatDataStruct) Snapshot() *StatDataStruct {
	//onlineIndex := s.GetOnlineIndex()
	onlineIndex := make(map[string]int)
	totalIndex := s.GetTotalIndex()
	data := make(map[string]*UserStat)
	//nowTs := time.Now().Unix()
	for k := range totalIndex {
		v := s.Get(k)
		if v != nil {
			//offlineAt := nowTs
			//if v.OfflineAt > 0 {
			//	offlineAt = v.OfflineAt
			//}

			data[k] = &UserStat{
				OnlineAt:    v.OnlineAt,
				OfflineAt:   v.OfflineAt,
				CalculateAt: v.CalculateAt,
				TotalOnline: v.TotalOnline,
				TodayOnline: v.TodayOnline,
			}
		}
	}

	return &StatDataStruct{
		Data:        data,
		OnlineIndex: onlineIndex,
	}
}

func (s *StatDataStruct) Get(k string) *UserStat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if x, found := s.Data[k]; found {
		return x
	}
	return nil
}

func (s *StatDataStruct) GetOnlineIndex() map[string]int {
	m := make(map[string]int)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.OnlineIndex {
		m[k] = 1
	}
	return m
}

func (s *StatDataStruct) GetTotalIndex() map[string]int {
	m := make(map[string]int)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.Data {
		m[k] = 1
	}
	return m
}

func (s *StatDataStruct) ClearData() {
	m := make(map[string]int)
	s.mu.RLock()
	for k := range s.Data {
		m[k] = 1
	}
	s.mu.RUnlock()

	for k := range m {
		s.clearOldData(k)
	}
}

func (s *StatDataStruct) Ticker() {
	m := make(map[string]int)
	s.mu.RLock()
	for k := range s.OnlineIndex {
		m[k] = 1
	}
	s.mu.RUnlock()

	for k := range m {
		s.Cal(k)
	}
}

func (s *StatDataStruct) clearOldData(k string) {
	v := s.Get(k)
	if v == nil {
		return
	}
	nowTs := time.Now().Unix()
	gap := nowTs - v.CalculateAt
	gap2 := nowTs - v.OnlineAt
	if gap2 >= clearGap*2 && gap >= clearGap*2 {
		s.mu.Lock()
		delete(s.Data, k)
		delete(s.OnlineIndex, k)
		s.mu.Unlock()
	}
	return
}

func (s *StatDataStruct) Cal(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowTs := time.Now().Unix()
	if x, found := s.Data[k]; found {
		if x.OfflineAt > x.OnlineAt || x.CalculateAt == 0 {
			return
		}

		gap := nowTs - x.CalculateAt
		if gap == 0 {
			return
		}

		x.TotalOnline += gap
		todayBeginning := GetTodayBeginning()
		if x.CalculateAt < todayBeginning {
			x.TodayOnline = nowTs - todayBeginning
		} else {
			x.TodayOnline += nowTs - x.CalculateAt
		}
		x.CalculateAt = nowTs
		s.Data[k] = x
	}
}

func (s *StatDataStruct) Online(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowTime := time.Now().Unix()
	if x, found := s.Data[k]; found {
		x.OnlineAt = nowTime
		x.CalculateAt = nowTime
		x.OfflineAt = 0
		s.Data[k] = x
	} else {

		userInfo := &UserStat{
			OnlineAt:    nowTime,
			OfflineAt:   0,
			CalculateAt: nowTime,
			TotalOnline: 0,
			TodayOnline: 0,
		}
		s.Data[k] = userInfo
	}
	s.OnlineIndex[k] = 1
}

func (s *StatDataStruct) Offline(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.OnlineIndex, k)
	nowTs := time.Now().Unix()
	if x, found := s.Data[k]; found {
		if x.OfflineAt > x.OnlineAt || x.CalculateAt == 0 {
			return
		}

		gap := nowTs - x.CalculateAt
		if gap > 0 {
			x.TotalOnline += gap
		}

		todayBeginning := GetTodayBeginning()
		if x.CalculateAt < todayBeginning {
			x.TodayOnline = nowTs - todayBeginning
		} else {
			x.TodayOnline += nowTs - x.CalculateAt
		}

		x.OfflineAt = nowTs
		x.CalculateAt = nowTs
		s.Data[k] = x
	}
}

func GetTodayBeginning() int64 {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, shanghai).Unix()
}

func GetTomorrowBeginning() int64 {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, shanghai).Unix()
}
