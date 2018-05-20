//
// Session Tracking
//

package sessions

import "sync"

type Bucket_t struct {
	mx sync.Mutex
	storage * Storage_t
}

func NewBucket(ttl int64, count int) (self * Bucket_t) {
	self = &Bucket_t{}
	self.storage = NewStorage(ttl, count)
	return
}

func (self * Bucket_t) Clear() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Clear()
}

func (self * Bucket_t) Flush(LastTs int64, keep int, evicted Evict) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.storage.Flush(LastTs, keep, evicted)
}

func (self * Bucket_t) Remove(Domain interface{}, UID interface{}, evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Remove(Domain, UID, evicted)
}

func (self * Bucket_t) ListFront(evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.ListFront(evicted)
}

func (self * Bucket_t) ListBack(evicted Evict) bool {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.ListBack(evicted)
}

func (self * Bucket_t) Update(Ts int64, Domain interface{}, UID interface{}, Data interface{}, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Update(Ts, Domain,UID, Data, evicted)
}

func (self * Bucket_t) Stat(Domain interface{}) Stat_t {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Stat(Domain)
}

func (self * Bucket_t) StatList() (res StatList_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.StatList()
}

func (self * Bucket_t) Size() (int, int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	return self.storage.Size()
}
