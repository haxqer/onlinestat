package main

import (
	"sync"
	"time"
)

type StatDataStruct struct {
	mu          sync.RWMutex
	data        map[string]*UserStat
	onlineIndex map[string]int
}

type UserStat struct {
	OnlineAt    int64
	OfflineAt   int64
	CalculateAt int64
	TotalOnline int64
	TodayOnline int64
}

var (
	shanghai   *time.Location
	StatData   *StatDataStruct
	calLimiter <-chan time.Time
)

func init() {
	var err error
	shanghai, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	StatData = &StatDataStruct{
		data:        make(map[string]*UserStat),
		onlineIndex: make(map[string]int),
	}

	calLimiter = time.After(2 * time.Second)
	tickCalData()
}

func tickCalData() {
	go func() {
		for {
			<-calLimiter
			StatData.Ticker()
		}
	}()
}

func (s *StatDataStruct) Get(k string) *UserStat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if x, found := s.data[k]; found {
		return x
	}
	return nil
}

func (s *StatDataStruct) GetOnlineIndex() map[string]int {
	m := make(map[string]int)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.onlineIndex {
		m[k] = 1
	}
	return m
}

func (s *StatDataStruct) GetTotalIndex() map[string]int {
	m := make(map[string]int)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.data {
		m[k] = 1
	}
	return m
}

func (s *StatDataStruct) Ticker() {
	m := make(map[string]int)
	s.mu.RLock()
	for k := range s.onlineIndex {
		m[k] = 1
	}
	s.mu.RUnlock()

	for k := range m {
		s.Cal(k)
	}
	calLimiter = time.After(2 * time.Second)
}

func (s *StatDataStruct) Cal(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowTs := time.Now().Unix()
	if x, found := s.data[k]; found {
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
		s.data[k] = x
	}
}

func (s *StatDataStruct) Online(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowTime := time.Now().Unix()
	if x, found := s.data[k]; found {
		x.OnlineAt = nowTime
		x.OfflineAt = 0
		s.data[k] = x
	} else {

		userInfo := &UserStat{
			OnlineAt:    nowTime,
			OfflineAt:   0,
			CalculateAt: nowTime,
			TotalOnline: 0,
			TodayOnline: 0,
		}
		s.data[k] = userInfo
	}
	s.onlineIndex[k] = 1
}

func (s *StatDataStruct) Offline(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.onlineIndex, k)
	nowTs := time.Now().Unix()
	if x, found := s.data[k]; found {
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
		s.data[k] = x
	}
}

func GetTodayBeginning() int64 {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, shanghai).Unix()
}
