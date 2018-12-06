//
// Session Tracking Logic
//

package sessions

import "github.com/ondi/go-cache"

type Key_t struct {
	Domain interface{}
	UID interface{}
}

type Data_t interface {
	Lock()
}

type Mapped_t struct {
	Hits int64
	LeftTs int64
	RightTs int64
	Data Data_t
}

type Value_t struct {
	Key_t
	Mapped_t
}

type Storage_t struct {
	cc * cache.Cache_t
	ttl int64
	count int
	deferred bool
	domains Domains
	new_uid_data func () Data_t
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

func NewStorage(ttl int64, count int, deferred bool, domains Domains, new_uid_data func () Data_t) (self * Storage_t) {
	self = &Storage_t{}
	self.cc = cache.New()
	if ttl <= 0 {
		ttl = 1 << 63 - 1
	}
	if count <= 0 {
		count = 1 << 63 - 1
	}
	self.ttl = ttl
	self.count = count
	self.deferred = deferred
	self.domains = domains
	self.new_uid_data = new_uid_data
	return
}

func (self * Storage_t) Clear() {
	self.cc = cache.New()
	self.domains.Clear()
}

func (self * Storage_t) Remove(Domain interface{}, UID interface{}, evicted Evict) bool {
	if it := self.cc.Find(Key_t{Domain: Domain, UID: UID}); it != self.cc.End() {
		self.remove(it, evicted)
		return true
	}
	return false
}

func (self * Storage_t) Flush(Ts int64, keep int, evicted Evict) {
	for it := self.cc.Back(); it != self.cc.End() && self.evict(it, Ts, keep, evicted); it = it.Prev() {}
}

func (self * Storage_t) remove(it * cache.Value_t, evicted Evict) {
	value := Value_t{Key_t: it.Key().(Key_t), Mapped_t: it.Mapped().(Mapped_t)}
	self.domains.Remove(value.Domain, value.Hits, value.RightTs - value.LeftTs)
	self.cc.Remove(value.Key_t)
	evicted.Evict(value)
}

func (self * Storage_t) evict(it * cache.Value_t, Ts int64, keep int, evicted Evict) bool {
	if self.cc.Size() > keep ||
		self.deferred == false && (Ts - it.Mapped().(Mapped_t).RightTs > self.ttl || it.Mapped().(Mapped_t).LeftTs - Ts > self.ttl) {
		self.remove(it, evicted)
		return true
	}
	return false
}

func (self * Storage_t) push_front(Ts int64, Domain interface{}, UID interface{}, evicted Evict) (it * cache.Value_t, Mapped Mapped_t, ok bool) {
	if it, ok = self.cc.PushFront(Key_t{Domain: Domain, UID: UID}, Mapped_t{}); ok {
		Mapped = Mapped_t{Hits: 1, LeftTs: Ts, RightTs: Ts, Data: self.new_uid_data()}
		self.domains.Add(Domain, Mapped.Data)
		it.Update(Mapped)
	} else {
		Mapped = it.Mapped().(Mapped_t)
	}
	return
}

func (self * Storage_t) Update(Ts int64, Domain interface{}, UID interface{}, evicted Evict) (Diff int64, Mapped Mapped_t) {
	var ok bool
	var it * cache.Value_t
	self.Flush(Ts, self.count, evicted)
	if it, Mapped, ok = self.push_front(Ts, Domain, UID, evicted); ok {
		Mapped.Data.Lock()
		return
	}
	if self.deferred && (Ts - Mapped.RightTs > self.ttl || Mapped.LeftTs - Ts > self.ttl) {
		self.remove(it, evicted)
		_, Mapped, _ = self.push_front(Ts, Domain, UID, evicted)
		Mapped.Data.Lock()
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
	self.domains.Update(Domain, Mapped.Hits, Diff)
	it.Update(Mapped)
	Mapped.Data.Lock()
	return
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

func (self * Storage_t) Size() int {
	return self.cc.Size()
}

func (self * Storage_t) DomainsSize() int {
	return self.domains.Size()
}

func (self * Storage_t) Stats() (stats Stats, ok bool) {
	stats, ok = self.domains.(Stats)
	return
}
