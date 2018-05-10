//
// Session Tracking
//

package sessions

import "sync"

import "github.com/ondi/go-cache"
// import "github.com/ondi/go-log"

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

type Bucket_t struct {
	mx sync.Mutex
	cc * cache.Cache
	stats map[ID64_t]*Stat_t
}

type Session_t struct {
	shards uint64
	ttl int64
	count int
	bucket []Bucket_t
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

func (self * Bucket_t) __evict_last(LastTs int64, keep int, ttl int64, evicted Evict) bool {
	if it := self.cc.Back(); it != self.cc.End() && (LastTs - it.Mapped().(Mapped_t).LastTs > ttl || self.cc.Size() > keep) {
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

func (self * Bucket_t) Flush(LastTs int64, Keep int, TTL int64, evicted Evict) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for self.__evict_last(LastTs, Keep, TTL, evicted) {}
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

func (self * Bucket_t) Update(Ts int64, Domain ID64_t, UID ID64_t, Data interface{}, count int, ttl int64, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for self.__evict_last(Ts, count, ttl, evicted) {}
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

func (self * Bucket_t) StatAll(res map[ID64_t]Stat_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.stats {
		res[k] = *v
	}
}

func (self * Bucket_t) Size() (int, int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.cc.Size(), len(self.stats)
}

func New(shards uint64, ttl int64, count int) (self * Session_t) {
	self = &Session_t{}
	if ttl <= 0 {
		ttl = 1 << 63 - 1
	}
	if count <= 0 {
		count = 1 << 63 - 1
	}
	self.ttl = ttl
	self.count = count
	self.shards = shards
	self.bucket = make([]Bucket_t, shards)
	self.Clear()
	return
}

func (self * Session_t) get_bucket(Domain ID64_t) uint64 {
	return Domain.Sum64() % self.shards
}

func (self * Session_t) Clear() {
	for i := uint64(0); i < self.shards; i++ {
		self.bucket[i].Clear()
	}
}

func (self * Session_t) Flush(LastTs int64, Keep int, evicted Evict) {
	for i := uint64(0); i < self.shards; i++ {
		self.bucket[i].Flush(LastTs, Keep, self.ttl, evicted)
	}
}

func (self * Session_t) Remove(Domain ID64_t, UID ID64_t, evicted Evict) bool {
	i := self.get_bucket(Domain)
	return self.bucket[i].Remove(Domain, UID, evicted)
}

func (self * Session_t) ListFront(evicted Evict) {
	for i := uint64(0); i < self.shards; i++ {
		if self.bucket[i].ListFront(evicted) == false {
			return
		}
	}
}

func (self * Session_t) ListBack(evicted Evict) {
	for i := uint64(0); i < self.shards; i++ {
		if self.bucket[i].ListBack(evicted) == false {
			return
		}
	}
}

func (self * Session_t) Update(Ts int64, Domain ID64_t, UID ID64_t, Data interface{}, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	i := self.get_bucket(Domain)
	return self.bucket[i].Update(Ts, Domain, UID, Data, self.count, self.ttl, evicted)
}

func (self * Session_t) Stat(Domain ID64_t) (stat Stat_t) {
	i := self.get_bucket(Domain)
	return self.bucket[i].Stat(Domain)
}

func (self * Session_t) StatAll() (res map[ID64_t]Stat_t) {
	res = map[ID64_t]Stat_t{}
	for i := uint64(0); i < self.shards; i++ {
		self.bucket[i].StatAll(res)
	}
	return
}

func (self * Session_t) SizeBuckets() (res [][3]int) {
	for i := uint64(0); i < self.shards; i++ {
		a, b := self.bucket[i].Size()
		res = append(res, [3]int{1, a, b})
	}
	return
}

func (self * Session_t) Size() (res [][3]int) {
	var temp [3]int
	for i := uint64(0); i < self.shards; i++ {
		a, b := self.bucket[i].Size()
		temp[1] += a
		temp[2] += b
	}
	temp[0] = int(self.shards)
	return [][3]int{temp}
}

func (self * Session_t) TTL() int64 {
	return self.ttl
}

func (self * Session_t) Count() int {
	return self.count
}
