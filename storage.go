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

func NewStorage(ttl int64, limit int, domains Domains) (self * Storage_t) {
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
	return
}

func (self * Storage_t) Clear() {
	self.c.Clear()
	self.domains.Clear()
}

func (self * Storage_t) Remove(Domain interface{}, UID interface{}, evicted Evict) bool {
	if it := self.c.Find(Key_t{Domain: Domain, UID: UID}); it != self.c.End() {
		self.remove(it, evicted)
		return true
	}
	return false
}

func (self * Storage_t) Flush(Ts int64, keep int, evicted Evict) {
	for it := self.c.Back(); it != self.c.End() && self.evict(it, Ts, keep, evicted); it = it.Prev() {}
}

func (self * Storage_t) remove(it * cache.Value_t, evicted Evict) {
	value := Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Value().(Mapped_t)}
	self.domains.RemoveSession(value.Domain, value.Hits, value.RightTs - value.LeftTs)
	self.c.Remove(value.Key_t)
	evicted.Evict(value)
}

func (self * Storage_t) evict(it * cache.Value_t, Ts int64, keep int, evicted Evict) bool {
	if self.c.Size() > keep || Ts - it.Value().(Mapped_t).RightTs > self.ttl || it.Value().(Mapped_t).LeftTs - Ts > self.ttl {
		self.remove(it, evicted)
		return true
	}
	return false
}

func (self * Storage_t) push_front(Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}) (it * cache.Value_t, Mapped Mapped_t, ok bool) {
	if it, ok = self.c.PushFront(Key_t{Domain: Domain, UID: UID}, nil); ok {
		Mapped = Mapped_t{Hits: 1, LeftTs: Ts, RightTs: Ts, Data: NewData()}
		self.domains.AddSession(Domain, Mapped.Data)
		it.Update(Mapped)
	} else {
		Mapped = it.Value().(Mapped_t)
	}
	return
}

func (self * Storage_t) Update(Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}, evicted Evict) (Diff int64, Mapped Mapped_t) {
	var ok bool
	var it * cache.Value_t
	if it, Mapped, ok = self.push_front(Ts, Domain, UID, NewData); ok {
		self.Flush(Ts, self.limit, evicted)
		return
	}
	if Ts - Mapped.RightTs > self.ttl || Mapped.LeftTs - Ts > self.ttl {
		self.remove(it, evicted)
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

func (self * Storage_t) ListFront(evicted Evict) bool {
	for it := self.c.Front(); it != self.c.End(); it = it.Next() {
		if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Value().(Mapped_t)}) == false {
			return false
		}
	}
	return true
}

func (self * Storage_t) ListBack(evicted Evict) bool {
	for it := self.c.Back(); it != self.c.End(); it = it.Prev() {
		if evicted.Evict(Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Value().(Mapped_t)}) == false {
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
