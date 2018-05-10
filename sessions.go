//
// Session Tracking
//

package sessions

type ID64_t interface {
	Sum64() uint64
}

type Session_t struct {
	shards uint64
	bucket []*Bucket_t
}

func New(shards uint64, ttl int64, count int) (self * Session_t) {
	self = &Session_t{}
	if ttl <= 0 {
		ttl = 1 << 63 - 1
	}
	if count <= 0 {
		count = 1 << 63 - 1
	}
	self.shards = shards
	for i := uint64(0); i < shards; i++ {
		self.bucket = append(self.bucket, NewBucket(ttl, count))
	}
	return
}

func (self * Session_t) get_bucket(Domain ID64_t) uint64 {
	return Domain.Sum64() % self.shards
}

func (self * Session_t) Clear() {
	for _, b := range self.bucket {
		b.Clear()
	}
}

func (self * Session_t) Flush(LastTs int64, keep int, evicted Evict) {
	for _, b := range self.bucket {
		b.Flush(LastTs, keep, evicted)
	}
}

func (self * Session_t) Remove(Domain ID64_t, UID interface{}, evicted Evict) bool {
	i := self.get_bucket(Domain)
	return self.bucket[i].Remove(Domain, UID, evicted)
}

func (self * Session_t) ListFront(evicted Evict) {
	for _, b := range self.bucket {
		if b.ListFront(evicted) == false {
			return
		}
	}
}

func (self * Session_t) ListBack(evicted Evict) {
	for _, b := range self.bucket {
		if b.ListBack(evicted) == false {
			return
		}
	}
}

func (self * Session_t) Update(Ts int64, Domain ID64_t, UID interface{}, Data interface{}, evicted Evict) (LastTs int64, Diff int64, Mapped Mapped_t) {
	i := self.get_bucket(Domain)
	return self.bucket[i].Update(Ts, Domain, UID, Data, evicted)
}

func (self * Session_t) Stat(Domain ID64_t) (stat Stat_t) {
	i := self.get_bucket(Domain)
	return self.bucket[i].Stat(Domain)
}

func (self * Session_t) StatList() (res StatList_t) {
	for _, b := range self.bucket {
		res = append(res, b.StatList()...)
	}
	return
}

func (self * Session_t) SizeBuckets() (res [][3]int) {
	for _, b := range self.bucket {
		x, y := b.Size()
		res = append(res, [3]int{1, x, y})
	}
	return
}

func (self * Session_t) Size() (res [][3]int) {
	var temp [3]int
	for _, b := range self.bucket {
		x, y := b.Size()
		temp[1] += x
		temp[2] += y
	}
	temp[0] = int(self.shards)
	return [][3]int{temp}
}
