//
// Session Tracking Synchronized
//

package sessions

import "sync"

type Session_t struct {
	mx sync.Mutex
	storage * Storage_t
}

func NewSession(ttl int64, limit int, domains Domains, evict Evict) (self * Session_t) {
	self = &Session_t{}
	self.storage = NewStorage(ttl, limit, domains, evict)
	return
}

func (self * Session_t) Clear() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Clear()
}

func (self * Session_t) Flush(Ts int64, keep int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Flush(Ts, keep)
}

func (self * Session_t) Remove(Domain interface{}, UID interface{}) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Remove(Domain, UID)
}

func (self * Session_t) ListFront(list Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.ListFront(list)
}

func (self * Session_t) ListBack(list Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.ListBack(list)
}

func (self * Session_t) Update(Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}) (Diff int64, Mapped Mapped_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	Diff, Mapped = self.storage.Update(Ts, Domain, UID, NewData)
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
