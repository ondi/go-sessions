//
// Session Tracking
//

package sessions

import "github.com/ondi/go-cache"

type Key_t struct {
	Domain interface{}
	UID interface{}
}

type Mapped_t struct {
	Hits int64
	Duration int64
	FirstTs int64
	LastTs int64
	FirstData interface{}
	LastData interface{}
}

type Value_t struct {
	Key_t
	Mapped_t
}

type Stat_t struct {
	Hits int64
	Sessions int64
	Bounces int64
	Duration int64
}

type StatRow_t struct {
	Domain interface{}
	Stat Stat_t
}

type StatList_t []StatRow_t

func (self StatList_t) Len() int {
	return len(self)
}

func (self StatList_t) Swap(i int, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self StatList_t) Less(i int, j int) bool {
	return self[i].Stat.Hits < self[j].Stat.Hits
}

type Storage_t struct {
	cc * cache.Cache
	stats map[interface{}]*Stat_t
	ttl int64
	count int
}

type Evict interface {
	Evict(Value_t) bool
}

type Evict_t []Value_t

func (self * Evict_t) Evict(value Value_t) bool {
	*self = append(*self, value)
	return true
}

type Drop_t struct {}

func (Drop_t) Evict(Value_t) bool {
	return true
}

func NewStorage(ttl int64, count int) (self * Storage_t) {
	self = &Storage_t{}
	self.cc = cache.New()
	self.stats = make(map[interface{}]*Stat_t)
	self.ttl = ttl
	self.count = count
	return
}

func (self * Storage_t) remove(it * cache.Value_t, evicted Evict) {
	value := Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}
	stat := self.stats[value.Domain]
	if stat.Sessions > 1 {
		stat.Sessions--
		if value.Hits == 1 {
			stat.Bounces--
		}
		stat.Hits -= value.Hits
		stat.Duration -= value.Duration
	} else {
		delete(self.stats, value.Domain)
	}
	self.cc.Remove(value.Key_t)
	evicted.Evict(value)
}

func (self * Storage_t) evict_last(LastTs int64, keep int, evicted Evict) bool {
	if it := self.cc.Back(); it != self.cc.End() && (LastTs - it.Mapped().(Mapped_t).LastTs > self.ttl || self.cc.Size() > keep) {
		self.remove(it, evicted)
		return true
	}
	return false
}

func (self * Storage_t) Clear() {
	self.cc = cache.New()
	self.stats = make(map[interface{}]*Stat_t)
}

func (self * Storage_t) Flush(LastTs int64, keep int, evicted Evict) {
	for self.evict_last(LastTs, keep, evicted) {}
}

func (self * Storage_t) Remove(Domain interface{}, UID interface{}, evicted Evict) bool {
	if it := self.cc.Find(Key_t{Domain: Domain, UID: UID}); it != self.cc.End() {
		self.remove(it, evicted)
		return true
	}
	return false
}

func (self * Storage_t) ListFront(evicted Evict) bool {
	for it := self.cc.Front(); it != self.cc.End(); it = it.Next() {
		if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}) == false {
			return false
		}
	}
	return true
}

func (self * Storage_t) ListBack(evicted Evict) bool {
	for it := self.cc.Back(); it != self.cc.End(); it = it.Prev() {
		if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}) == false {
			return false
		}
	}
	return true
}

func (self * Storage_t) Update(Ts int64, Domain interface{}, UID interface{}, Data interface{}, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	for self.evict_last(Ts, self.count, evicted) {}
	key := Key_t{Domain: Domain, UID: UID}
	Mapped = Mapped_t{FirstData: Data, LastData: Data, Hits: 1, Duration: 0, FirstTs: Ts, LastTs: Ts}
	it, ok := self.cc.PushFront(key, Mapped)
	if ok {
		if stat, ok := self.stats[Domain]; ok {
			stat.Hits++
			stat.Sessions++
			stat.Bounces++
		} else {
			self.stats[Domain] = &Stat_t{Hits: 1, Sessions: 1, Bounces: 1, Duration: 0}
		}
		return
	}
	Mapped = it.Mapped().(Mapped_t)
	LastTs = Mapped.LastTs
	if Ts >= Mapped.LastTs {
		Diff = Ts - Mapped.LastTs
		Mapped.LastTs = Ts
		// Mapped.LastData = Data
	} else if Ts <= Mapped.FirstTs {
		Diff = Mapped.FirstTs - Ts
		Mapped.FirstTs = Ts
		// Mapped.FirstData = Data
	}
	if Data != nil {
		Mapped.LastData = Data
		if Mapped.FirstData == nil {
			Mapped.FirstData = Data
		}
	}
	Mapped.Hits++
	Mapped.Duration += Diff
	stat := self.stats[Domain]
	if Mapped.Hits == 2 {
		stat.Bounces--
	}
	stat.Hits++
	stat.Duration += Diff
	it.Update(Mapped)
	return
}

func (self * Storage_t) Stat(Domain interface{}) Stat_t {
	if res, ok := self.stats[Domain]; ok {
		return *res
	}
	return Stat_t{}
}

func (self * Storage_t) StatList() (res StatList_t) {
	for k, v := range self.stats {
		res = append(res, StatRow_t{k, *v})
	}
	return
}

func (self * Storage_t) Size() (int, int) {
	return len(self.stats), self.cc.Size()
}
