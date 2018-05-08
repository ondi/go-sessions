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

type bucket_t struct {
	mx sync.Mutex
	cx * cache.Cache
	stats map[ID64_t]*Stat_t
}

type Session_t struct {
	shards uint64
	ttl int64
	count int
	bucket []bucket_t
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
	self.bucket = make([]bucket_t, shards)
	self.Clear()
	return
}

func (self * Session_t) get_bucket(Domain ID64_t) uint64 {
	return Domain.Sum64() % self.shards
}

func (self * Session_t) Clear() {
	for i := uint64(0); i < self.shards; i++ {
		self.clear(i)
	}
}

func (self * Session_t) clear(i uint64) {
	self.bucket[i].mx.Lock()
	defer self.bucket[i].mx.Unlock()
	self.bucket[i].cx = cache.New()
	self.bucket[i].stats = make(map[ID64_t]*Stat_t)
}

func (self * Session_t) __remove(i uint64, it * cache.Value_t, evicted Evict) {
	value := Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}
	stat := self.bucket[i].stats[value.Domain]
	if stat.Sessions > 1 {
		stat.Sessions--
		if value.Hits == 1 {
			stat.Bounces--
		}
		stat.Hits -= value.Hits
		stat.Duration -= value.Duration
	} else {
		delete(self.bucket[i].stats, value.Domain)
	}
	self.bucket[i].cx.Remove(value.Key_t)
	evicted.Evict(value)
}

func (self * Session_t) __evict_last(i uint64, LastTs int64, Keep int, evicted Evict) bool {
	if it := self.bucket[i].cx.Back(); it != self.bucket[i].cx.End() && (LastTs - it.Mapped().(Mapped_t).LastTs > self.ttl || self.bucket[i].cx.Size() > Keep) {
		self.__remove(i, it, evicted)
		return true
	}
	return false
}

func (self * Session_t) Flush(LastTs int64, Keep int, evicted Evict) {
	for i := uint64(0); i < self.shards; i++ {
		self.bucket[i].mx.Lock()
		for self.__evict_last(i, LastTs, Keep, evicted) {}
		self.bucket[i].mx.Unlock()
	}
}

func (self * Session_t) Remove(Domain ID64_t, UID ID64_t, evicted Evict) bool {
	i := self.get_bucket(Domain)
	self.bucket[i].mx.Lock()
	defer self.bucket[i].mx.Unlock()
	if it := self.bucket[i].cx.Find(Key_t{Domain: Domain, UID: UID}); it != self.bucket[i].cx.End() {
		self.__remove(i, it, evicted)
		return true
	}
	return false
}

func (self * Session_t) ListFront(evicted Evict) {
	for i := uint64(0); i < self.shards; i++ {
		self.bucket[i].mx.Lock()
		for it := self.bucket[i].cx.Front(); it != self.bucket[i].cx.End(); it = it.Next() {
			if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}) == false {
				break
			}
		}
		self.bucket[i].mx.Unlock()
	}
}

func (self * Session_t) ListBack(evicted Evict) {
	for i := uint64(0); i < self.shards; i++ {
		self.bucket[i].mx.Lock()
		for it := self.bucket[i].cx.Back(); it != self.bucket[i].cx.End(); it = it.Prev() {
			if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}) == false {
				break
			}
		}
		self.bucket[i].mx.Unlock()
	}
}

func (self * Session_t) Update(Ts int64, Domain ID64_t, UID ID64_t, Data interface{}, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	i := self.get_bucket(Domain)
	self.bucket[i].mx.Lock()
	defer self.bucket[i].mx.Unlock()
	for self.__evict_last(i, Ts, self.count, evicted) {}
	key := Key_t{Domain: Domain, UID: UID}
	Mapped = Mapped_t{FirstData: Data, LastData: Data, Hits: 1, Duration: 0, FirstTs: Ts, LastTs: Ts}
	it, ok := self.bucket[i].cx.PushFront(key, Mapped)
	if ok {
		if stat, ok := self.bucket[i].stats[Domain]; ok {
			stat.Hits++
			stat.Sessions++
			stat.Bounces++
		} else {
			self.bucket[i].stats[Domain] = &Stat_t{Hits: 1, Sessions: 1, Bounces: 1, Duration: 0}
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
	stat := self.bucket[i].stats[Domain]
	if Mapped.Hits == 2 {
		stat.Bounces--
	}
	stat.Hits++
	stat.Duration += Diff
	it.Update(Mapped)
	return
}

func (self * Session_t) Stat(Domain ID64_t) (stat Stat_t) {
	i := self.get_bucket(Domain)
	self.bucket[i].mx.Lock()
	defer self.bucket[i].mx.Unlock()
	if res, ok := self.bucket[i].stats[Domain]; ok {
		stat = *res
	}
	return
}

func (self * Session_t) stat_all(i uint64, res map[ID64_t]Stat_t) () {
	self.bucket[i].mx.Lock()
	defer self.bucket[i].mx.Unlock()
	for k, v := range self.bucket[i].stats {
		res[k] = *v
	}
}

func (self * Session_t) StatAll() (res map[ID64_t]Stat_t) {
	res = map[ID64_t]Stat_t{}
	for i := uint64(0); i < self.shards; i++ {
		self.stat_all(i, res)
	}
	return
}

func (self * Session_t) size_bucket(i uint64, res * [3]int) {
	self.bucket[i].mx.Lock()
	defer self.bucket[i].mx.Unlock()
	res[0] += 1
	res[1] += self.bucket[i].cx.Size()
	res[2] += len(self.bucket[i].stats)
}

func (self * Session_t) SizeBuckets() (res [][3]int) {
	for i := uint64(0); i < self.shards; i++ {
		res = append(res, [3]int{})
		self.size_bucket(i, &res[i])
	}
	return
}

func (self * Session_t) Size() (res [][3]int) {
	res = append(res, [3]int{})
	for i := uint64(0); i < self.shards; i++ {
		self.size_bucket(i, &res[0])
	}
	return
}

func (self * Session_t) TTL() int64 {
	return self.ttl
}

func (self * Session_t) Count() int {
	return self.count
}
