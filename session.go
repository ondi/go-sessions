//
// Session Tracking Synchronized
//

package sessions

import "sync"

type Session_t struct {
	mx sync.Mutex
	storage * Storage_t
}

func NewSession(ttl int64, count int, deferred bool, data func () Data_t) (self * Session_t) {
	self = &Session_t{}
	self.storage = NewStorage(ttl, count, deferred, data)
	return
}

func (self * Session_t) Clear() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Clear()
}

func (self * Session_t) Flush(LastTs int64, keep int, evicted Evict) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Flush(LastTs, keep, evicted)
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

func (self * Session_t) Update(Ts int64, Domain interface{}, UID interface{}, evicted Evict) (Diff int64, Mapped Mapped_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Update(Ts, Domain,UID, evicted)
}

func (self * Session_t) Stat(Domain interface{}) Stat_t {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Stat(Domain)
}

func (self * Session_t) StatList() (res []StatList_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.StatList()
}

func (self * Session_t) Size() (int, int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.StatSize(), self.storage.Size()
}
