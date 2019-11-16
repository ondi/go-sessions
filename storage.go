//
// Session Tracking Logic
//

package sessions

import "github.com/ondi/go-cache"

type Key_t struct {
	Domain interface{}
	UID interface{}
}

type Mapped_t struct {
	Hits int64
	LeftTs int64
	RightTs int64
	Data interface{}
}

type Value_t struct {
	Key_t
	Mapped_t
}

type Storage_t struct {
	c * cache.Cache_t
	ttl int64
	limit int
	domains Domains
	evict Evict
}

type Evict func(interface{}) int

func Drop(interface{}) int {return 0}

func NewStorage(ttl int64, limit int, domains Domains, evict Evict) (self * Storage_t) {
	self = &Storage_t{}
	self.c = cache.New()
	if ttl <= 0 {
		ttl = 1 << 63 - 1
	}
	if limit <= 0 {
		limit = 1 << 63 - 1
	}
	self.ttl = ttl
	self.limit = limit
	if domains == nil {
		self.domains = NoDomains_t{}
	} else {
		self.domains = domains
	}
	self.evict = evict
	return
}

func (self * Storage_t) Clear() {
	self.c.Clear()
	self.domains.Clear()
}

func (self * Storage_t) remove(it * cache.Value_t) {
	value := Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Value().(Mapped_t)}
	self.domains.RemoveSession(value.Domain, value.Hits, value.RightTs - value.LeftTs)
	self.c.Remove(value.Key_t)
	self.evict(value)
}

func (self * Storage_t) flush(it * cache.Value_t, Ts int64, keep int) bool {
	if self.c.Size() > keep || Ts - it.Value().(Mapped_t).RightTs > self.ttl || it.Value().(Mapped_t).LeftTs - Ts > self.ttl {
		self.remove(it)
		return true
	}
	return false
}

func (self * Storage_t) push_front(Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}) (it * cache.Value_t, Mapped Mapped_t, ok bool) {
	if it, ok = self.c.PushFront(Key_t{Domain: Domain, UID: UID}, func() interface{} {return Mapped_t{Hits: 1, LeftTs: Ts, RightTs: Ts, Data: NewData()}}); ok {
		self.domains.AddSession(Domain, it.Value().(Mapped_t).Data)
	}
	Mapped = it.Value().(Mapped_t)
	return
}

func (self * Storage_t) Remove(Domain interface{}, UID interface{}) (ok bool) {
	var it * cache.Value_t
	if it, ok = self.c.Find(Key_t{Domain: Domain, UID: UID}); ok {
		self.remove(it)
	}
	return
}

func (self * Storage_t) Flush(Ts int64, keep int) {
	for it := self.c.Back(); it != self.c.End() && self.flush(it, Ts, keep); it = it.Prev() {}
}

func (self * Storage_t) Update(Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}) (Diff int64, Mapped Mapped_t) {
	var ok bool
	var it * cache.Value_t
	if it, Mapped, ok = self.push_front(Ts, Domain, UID, NewData); ok {
		self.Flush(Ts, self.limit)
		return
	}
	if Ts - Mapped.RightTs > self.ttl || Mapped.LeftTs - Ts > self.ttl {
		self.remove(it)
		_, Mapped, _ = self.push_front(Ts, Domain, UID, NewData)
		return
	}
	Mapped.Hits++
	if Ts > Mapped.RightTs {
		Diff = Ts - Mapped.RightTs
		Mapped.RightTs = Ts
	} else if Ts < Mapped.LeftTs {
		Diff = Mapped.LeftTs - Ts
		Mapped.LeftTs = Ts
	}
	it.Update(Mapped)
	self.domains.UpdateSession(Domain, Mapped.Hits, Diff)
	return
}

func (self * Storage_t) ListFront(evict Evict) bool {
	for it := self.c.Front(); it != self.c.End(); it = it.Next() {
		if evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Value().(Mapped_t)}) != 0 {
			return false
		}
	}
	return true
}

func (self * Storage_t) ListBack(evict Evict) bool {
	for it := self.c.Back(); it != self.c.End(); it = it.Prev() {
		if evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Value().(Mapped_t)}) != 0 {
			return false
		}
	}
	return true
}

func (self * Storage_t) Size() int {
	return self.c.Size()
}

func (self * Storage_t) Stats() (stats Stats, ok bool) {
	stats, ok = self.domains.(Stats)
	return
}
