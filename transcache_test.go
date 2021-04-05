/*
TransCache is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM GmbH. All Rights Reserved.
*/

package ltcache

import (
	"container/list"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestRemKey(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	tc.Set("t11_", "mm", "test", nil, true, "")
	if t1, ok := tc.Get("t11_", "mm"); !ok || t1 != "test" {
		t.Error("Error setting cache: ", ok, t1)
	}
	tc.Remove("t11_", "mm", true, "")
	if t1, ok := tc.Get("t11_", "mm"); ok || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestTransaction(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	transID := tc.BeginTransaction()
	tc.Set("mmm_", "t11", "test", nil, false, transID)
	if t1, ok := tc.Get("mmm_", "t11"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	tc.Set("mmm_", "t12", "test", nil, false, transID)
	tc.Remove("mmm_", "t11", false, transID)
	if _, hasTransID := tc.transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	tc.CommitTransaction(transID)
	if t1, ok := tc.Get("mmm_", "t12"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := tc.Get("mmm_", "t11"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := tc.transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}

}

func TestTransactionRemove(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	transID := tc.BeginTransaction()
	tc.Set("t21_", "mm", "test", nil, false, transID)
	tc.Set("t21_", "nn", "test", nil, false, transID)
	tc.Remove("t21_", "mm", false, transID)
	if _, hasTransID := tc.transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	tc.CommitTransaction(transID)
	if t1, ok := tc.Get("t21_", "mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := tc.Get("t21_", "nn"); !ok || t1 != "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := tc.transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}
}

func TestTransactionRemoveGroup(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	transID := tc.BeginTransaction()
	tc.Set("t21_", "mm", "test", []string{"grp1"}, false, transID)
	tc.Set("t21_", "nn", "test", []string{"grp1"}, false, transID)
	tc.RemoveGroup("t21_", "grp1", false, transID)
	if _, hasTransID := tc.transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	tc.CommitTransaction(transID)
	if t1, ok := tc.Get("t21_", "mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := tc.Get("t21_", "nn"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := tc.transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}
}

func TestTransactionRollback(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	transID := tc.BeginTransaction()
	tc.Set("aaa_", "t31", "test", nil, false, transID)
	if t1, ok := tc.Get("aaa_", "t31"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	tc.Set("aaa_", "t32", "test", nil, false, transID)
	if _, hasTransID := tc.transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	tc.RollbackTransaction(transID)
	if t1, ok := tc.Get("aaa_", "t32"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := tc.Get("aaa_", "t31"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := tc.transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}
}

func TestTransactionRemBefore(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	transID := tc.BeginTransaction()
	tc.Remove("t41_", "mm", false, transID)
	tc.Remove("t41_", "nn", false, transID)
	tc.Set("t41_", "mm", "test", nil, false, transID)
	tc.Set("t41_", "nn", "test", nil, false, transID)
	tc.CommitTransaction(transID)
	if t1, ok := tc.Get("t41_", "mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := tc.Get("t41_", "nn"); !ok || t1 != "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTCGetGroupItems(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	tc.Set("xxx_", "t1", "test", []string{"grp1"}, true, "")
	tc.Set("xxx_", "t2", "test", []string{"grp1"}, true, "")
	if grpItms := tc.GetGroupItems("xxx_", "grp1"); len(grpItms) != 2 {
		t.Errorf("Received group items: %+v", grpItms)
	}
	if grpItms := tc.GetGroupItems("xxx_", "nonexsitent"); grpItms != nil {
		t.Errorf("Received group items: %+v", grpItms)
	}
}

func TestRemGroup(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	tc.Set("xxx_", "t1", "test", []string{"grp1"}, true, "")
	tc.Set("xxx_", "t2", "test", []string{"grp1"}, true, "")
	tc.RemoveGroup("xxx_", "grp1", true, "")
	_, okt1 := tc.Get("xxx_", "t1")
	_, okt2 := tc.Get("xxx_", "t2")
	if okt1 || okt2 {
		t.Error("Error removing prefix: ", okt1, okt2)
	}
}

func TestCacheCount(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{
		"dst_": &CacheConfig{MaxItems: -1},
		"rpf_": &CacheConfig{MaxItems: -1}})
	tc.Set("dst_", "A1", "1", nil, true, "")
	tc.Set("dst_", "A2", "2", nil, true, "")
	tc.Set("rpf_", "A3", "3", nil, true, "")
	tc.Set("dst_", "A4", "4", nil, true, "")
	tc.Set("dst_", "A5", "5", nil, true, "")
	if itms := tc.GetItemIDs("dst_", ""); len(itms) != 4 {
		t.Errorf("Error getting item ids: %+v", itms)
	}
}

func TestCacheGetStats(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{
		"part1": &CacheConfig{MaxItems: -1},
		"part2": &CacheConfig{MaxItems: -1}})
	testCIs := []*cachedItem{
		&cachedItem{itemID: "_1_", value: "one"},
		&cachedItem{itemID: "_2_", value: "two", groupIDs: []string{"grp1"}},
		&cachedItem{itemID: "_3_", value: "three", groupIDs: []string{"grp1", "grp2"}},
		&cachedItem{itemID: "_4_", value: "four", groupIDs: []string{"grp1", "grp2", "grp3"}},
		&cachedItem{itemID: "_5_", value: "five", groupIDs: []string{"grp4"}},
	}
	for _, ci := range testCIs {
		tc.Set("part1", ci.itemID, ci.value, ci.groupIDs, true, "")
	}
	for _, ci := range testCIs[:4] {
		tc.Set("part2", ci.itemID, ci.value, ci.groupIDs, true, "")
	}
	eCs := map[string]*CacheStats{
		"part1": &CacheStats{Items: 5, Groups: 4},
		"part2": &CacheStats{Items: 4, Groups: 3},
	}
	if cs := tc.GetCacheStats(nil); reflect.DeepEqual(eCs, cs) {
		t.Errorf("expecting: %+v, received: %+v", eCs, cs)
	}
}

// Try concurrent read/write of the cache
func TestCacheConcurrent(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{
		"dst_": &CacheConfig{MaxItems: -1},
		"rpf_": &CacheConfig{MaxItems: -1}})
	s := &struct{ Prefix string }{Prefix: "+49"}
	tc.Set("dst_", "DE", s, nil, true, "")
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			tc.Get("dst_", "DE")
			wg.Done()
		}()
	}
	s.Prefix = "+491"
	wg.Wait()
}

type TenantID struct {
	Tenant string
	ID     string
}

func (tID *TenantID) Clone() (interface{}, error) {
	tClone := new(TenantID)
	*tClone = *tID
	return tClone, nil
}

func TestGetClone(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	a := &TenantID{Tenant: "cgrates.org", ID: "ID#1"}
	tc.Set("t11_", "mm", a, nil, true, "")
	if t1, ok := tc.Get("t11_", "mm"); !ok {
		t.Error("Error setting cache: ", ok, t1)
	}
	if x, err := tc.GetCloned("t11_", "mm"); err != nil {
		t.Error(err)
	} else {
		tcCloned := x.(*TenantID)
		if !reflect.DeepEqual(tcCloned, a) {
			t.Errorf("Expecting: %+v, received: %+v", a, tcCloned)
		}
		a.ID = "ID#2"
		if reflect.DeepEqual(tcCloned, a) {
			t.Errorf("Expecting: %+v, received: %+v", a, tcCloned)
		}
	}
}

func TestGetClone2(t *testing.T) {
	tc := NewTransCache(map[string]*CacheConfig{})
	tc.Set("t11_", "mm", nil, nil, true, "")
	if x, err := tc.GetCloned("t11_", "mm"); err != nil {
		t.Error(err)
	} else if x != nil {
		t.Errorf("Expecting: nil, received: %+v", x)
	}
}

func TestTranscacheHasGroup(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {},
		},
	}
	chID := "testChID"
	grpID := "testGroupID"

	exp := false
	rcv := tc.HasGroup(chID, grpID)

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	tc.cache[chID].groups = map[string]map[string]struct{}{
		"testGroupID": make(map[string]struct{}),
	}

	exp = true
	rcv = tc.HasGroup(chID, grpID)

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestTranscacheGetGroupItemIDs(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				groups: map[string]map[string]struct{}{
					"testGroupID": {
						"testField1": struct{}{},
						"testField2": struct{}{},
					},
				},
			},
		},
	}
	chID := "testChID"
	grpID := "testGroupID"

	exp := []string{"testField1", "testField2"}
	rcv := tc.GetGroupItemIDs(chID, grpID)

	if len(exp) != len(rcv) {
		t.Fatalf("\nexpected slice length: <%+v>, \nreceived slice length: <%+v>",
			len(exp), len(rcv))
	}

	diff := make(map[string]int, len(exp))

	for _, valRcv := range rcv {
		diff[valRcv]++
	}
	for _, valExp := range exp {
		if _, ok := diff[valExp]; !ok {
			t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
		}
		diff[valExp] -= 1
		if diff[valExp] == 0 {
			delete(diff, valExp)
		}
	}
	if len(diff) != 0 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestTranscacheClearSpecific(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID1": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache: map[string]*cachedItem{
					"item1": {},
					"item2": {},
				},
			},
			"testChID2": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache: map[string]*cachedItem{
					"item1": {},
					"item2": {},
					"item3": {},
				},
			},
		},
	}
	chIDs := []string{"testChID2"}

	exp := &TransCache{
		cache: map[string]*Cache{
			"testChID1": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache: map[string]*cachedItem{
					"item1": {},
					"item2": {},
				},
			},
			"testChID2": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache:  map[string]*cachedItem{},
			},
		},
	}

	tc.Clear(chIDs)

	for key, value := range tc.cache {
		if len(value.cache) != len(exp.cache[key].cache) {
			t.Errorf("\nKey: <%+v>\nexpected nr of items: <%+v>, \nreceived nr of items: <%+v>",
				key, len(exp.cache[key].cache), len(value.cache))
		}
	}
}

func TestTranscacheClearAll(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID1": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache: map[string]*cachedItem{
					"item1": {},
					"item2": {},
				},
			},
			"testChID2": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache: map[string]*cachedItem{
					"item1": {},
					"item2": {},
					"item3": {},
				},
			},
		},
	}
	var chIDs []string

	exp := &TransCache{
		cache: map[string]*Cache{
			"testChID1": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache:  map[string]*cachedItem{},
			},
			"testChID2": {
				lruIdx: list.New(),
				ttlIdx: list.New(),
				cache:  map[string]*cachedItem{},
			},
		},
	}

	tc.Clear(chIDs)

	for key, value := range tc.cache {
		if len(value.cache) != len(exp.cache[key].cache) {
			t.Errorf("\nKey: <%+v>\nexpected nr of items: <%+v>, \nreceived nr of items: <%+v>",
				key, len(exp.cache[key].cache), len(value.cache))
		}
	}
}

func TestTranscacheGetItemExpiryTime(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				cache: map[string]*cachedItem{
					"testItemID": {
						expiryTime: time.Time{},
					},
				},
			},
		},
	}
	chID := "testChID"
	itmID := "testItemID"

	var exp time.Time
	expok := true
	rcv, ok := tc.GetItemExpiryTime(chID, itmID)

	if ok != expok {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", expok, ok)
	}

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestTranscacheHasItem(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				cache: map[string]*cachedItem{
					"testItemID": {},
				},
			},
		},
	}
	chID := "testChID"
	itmID := "testItemID"

	exp := true
	rcv := tc.HasItem(chID, itmID)

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestTranscacheNoItem(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				cache: map[string]*cachedItem{
					"otherItem": {},
				},
			},
		},
	}
	chID := "testChID"
	itmID := "testItemID"

	exp := false
	rcv := tc.HasItem(chID, itmID)

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestTranscacheGetClonedNotFound(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				cache: map[string]*cachedItem{
					"otherItem": {},
				},
			},
		},
	}
	chID := "testChID"
	itmID := "testItemID"

	experr := ErrNotFound
	rcv, err := tc.GetCloned(chID, itmID)

	if rcv != nil {
		t.Fatalf("\nexpected nil, \nreceived: <%+v>", rcv)
	}

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestTranscacheGetClonedNotClonable(t *testing.T) {
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				lruIdx: list.New(),
				lruRefs: map[string]*list.Element{
					"testItemID": {},
				},
				cache: map[string]*cachedItem{
					"testItemID": {
						value: 3,
					},
				},
			},
		},
	}
	chID := "testChID"
	itmID := "testItemID"

	experr := ErrNotClonable
	rcv, err := tc.GetCloned(chID, itmID)

	if rcv != nil {
		t.Fatalf("\nexpected nil, \nreceived: <%+v>", rcv)
	}

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

type clonerMock struct {
	testcase string
}

func (cM *clonerMock) Clone() (interface{}, error) {
	switch cM.testcase {
	case "clone error":
		err := fmt.Errorf("clone mock error")
		return nil, err
	}
	return nil, nil
}
func TestTranscacheGetClonedCloneError(t *testing.T) {
	cloner := &clonerMock{
		testcase: "clone error",
	}
	tc := &TransCache{
		cache: map[string]*Cache{
			"testChID": {
				lruIdx: list.New(),
				lruRefs: map[string]*list.Element{
					"testItemID": {},
				},
				cache: map[string]*cachedItem{
					"testItemID": {
						value: cloner,
					},
				},
			},
		},
	}
	chID := "testChID"
	itmID := "testItemID"

	experr := "clone mock error"
	rcv, err := tc.GetCloned(chID, itmID)

	if rcv != nil {
		t.Fatalf("\nexpected nil, \nreceived: <%+v>", rcv)
	}

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

//BenchmarkSet            	 3000000	       469 ns/op
func BenchmarkSet(b *testing.B) {
	cacheItems := [][]string{
		[]string{"aaa_", "1", "1"},
		[]string{"aaa_", "2", "1"},
		[]string{"aaa_", "3", "1"},
		[]string{"aaa_", "4", "1"},
		[]string{"aaa_", "5", "1"},
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(cacheItems)-1 // so we can have random index
	tc := NewTransCache(map[string]*CacheConfig{})
	for n := 0; n < b.N; n++ {
		ci := cacheItems[rand.Intn(max-min)+min]
		tc.Set(ci[0], ci[1], ci[2], nil, false, "")
	}
}

// BenchmarkSetWithGroups  	 3000000	       591 ns/op
func BenchmarkSetWithGroups(b *testing.B) {
	cacheItems := [][]string{
		[]string{"aaa_", "1", "1"},
		[]string{"aaa_", "2", "1"},
		[]string{"aaa_", "3", "1"},
		[]string{"aaa_", "4", "1"},
		[]string{"aaa_", "5", "1"},
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(cacheItems)-1 // so we can have random index
	tc := NewTransCache(map[string]*CacheConfig{})
	for n := 0; n < b.N; n++ {
		ci := cacheItems[rand.Intn(max-min)+min]
		tc.Set(ci[0], ci[1], ci[2], []string{"grp1", "grp2"}, false, "")
	}
}

// BenchmarkGet            	10000000	       163 ns/op
func BenchmarkGet(b *testing.B) {
	cacheItems := [][]string{
		[]string{"aaa_", "1", "1"},
		[]string{"aaa_", "2", "1"},
		[]string{"aaa_", "3", "1"},
		[]string{"aaa_", "4", "1"},
		[]string{"aaa_", "5", "1"},
	}
	tc := NewTransCache(map[string]*CacheConfig{})
	for _, ci := range cacheItems {
		tc.Set(ci[0], ci[1], ci[2], nil, false, "")
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(cacheItems)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		tc.Get("aaa_", cacheItems[rand.Intn(max-min)+min][0])
	}
}
