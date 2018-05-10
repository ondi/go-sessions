//
// Session Tracking
//

package sessions

import "sync"

import "github.com/ondi/go-cache"

type ID64_t interface {
	Sum64() uint64
}

type Key_t struct {
	Domain ID64_t
	UID ID64_t
}

type Mapped_t struct {
	FirstData interface{}
	LastData interface{}
	Hits int64
	Duration int64
	FirstTs int64
	LastTs int64
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

type StatList_t struct {
	Domain ID64_t
	Stat Stat_t
}

type Bucket_t struct {
	mx sync.Mutex
	cc * cache.Cache
	stats map[ID64_t]*Stat_t
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

func NewBucket(ttl int64, count int) (self * Bucket_t) {
	self = &Bucket_t{}
	self.cc = cache.New()
	self.stats = make(map[ID64_t]*Stat_t)
	self.ttl = ttl
	self.count = count
	return
}

func (self * Bucket_t) __remove(it * cache.Value_t, evicted Evict) {
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

func (self * Bucket_t) __evict_last(LastTs int64, keep int, evicted Evict) bool {
	if it := self.cc.Back(); it != self.cc.End() && (LastTs - it.Mapped().(Mapped_t).LastTs > self.ttl || self.cc.Size() > keep) {
		self.__remove(it, evicted)
		return true
	}
	return false
}

func (self * Bucket_t) Clear() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.cc = cache.New()
	self.stats = make(map[ID64_t]*Stat_t)
}

func (self * Bucket_t) Flush(LastTs int64, keep int, evicted Evict) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for self.__evict_last(LastTs, keep, evicted) {}
}

func (self * Bucket_t) Remove(Domain ID64_t, UID ID64_t, evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	if it := self.cc.Find(Key_t{Domain: Domain, UID: UID}); it != self.cc.End() {
		self.__remove(it, evicted)
		return true
	}
	return false
}

func (self * Bucket_t) ListFront(evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	for it := self.cc.Front(); it != self.cc.End(); it = it.Next() {
		if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}) == false {
			return false
		}
	}
	return true
}

func (self * Bucket_t) ListBack(evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	for it := self.cc.Back(); it != self.cc.End(); it = it.Prev() {
		if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}) == false {
			return false
		}
	}
	return true
}

func (self * Bucket_t) Update(Ts int64, Domain ID64_t, UID ID64_t, Data interface{}, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for self.__evict_last(Ts, self.count, evicted) {}
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

func (self * Bucket_t) Stat(Domain ID64_t) Stat_t {
	self.mx.Lock()
	defer self.mx.Unlock()
	if res, ok := self.stats[Domain]; ok {
		return *res
	}
	return Stat_t{}
}

func (self * Bucket_t) StatList() (res []StatList_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.stats {
		res = append(res, StatList_t{k, *v})
	}
	return
}

func (self * Bucket_t) Size() (int, int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return len(self.stats), self.cc.Size()
}
