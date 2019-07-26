//
// Session Tracking Synchronized
//

package sessions

import "sync"

type Session_t struct {
	mx sync.Mutex
	storage * Storage_t
}

func NewSession(ttl int64, limit int, domains Domains) (self * Session_t) {
	self = &Session_t{}
	self.storage = NewStorage(ttl, limit, domains)
	return
}

func (self * Session_t) Clear() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Clear()
}

func (self * Session_t) Flush(Ts int64, keep int, evicted Evict) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Flush(Ts, keep, evicted)
}

func (self * Session_t) Remove(Domain interface{}, UID interface{}, evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Remove(Domain, UID, evicted)
}

func (self * Session_t) ListFront(evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.ListFront(evicted)
}

func (self * Session_t) ListBack(evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.ListBack(evicted)
}

func (self * Session_t) Update(Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}, evicted Evict) (Diff int64, Mapped Mapped_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	Diff, Mapped = self.storage.Update(Ts, Domain, UID, NewData, evicted)
	return
}

func (self * Session_t) Size() (int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Size()
}

func (self * Session_t) Stat(Domain interface{}) Stat_t {
	self.mx.Lock()
	defer self.mx.Unlock()
	if s, ok := self.storage.Stats(); ok {
		return s.Stat(Domain)
	}
	return Stat_t{}
}

func (self * Session_t) StatList() map[interface{}]Stat_t {
	self.mx.Lock()
	defer self.mx.Unlock()
	if s, ok := self.storage.Stats(); ok {
		return s.StatList()
	}
	return nil
}

func (self * Session_t) DomainSize() (int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if s, ok := self.storage.Stats(); ok {
		return s.Size()
	}
	return 0
}
