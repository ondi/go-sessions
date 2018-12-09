//
// Session Tracking Shards
//

package sessions

type Sessions_t struct {
	shards uint64
	bucket []*Session_t
}

func NewSessions(shards uint64, ttl int64, count int, deferred bool, domains Domains) (self * Sessions_t) {
	self = &Sessions_t{}
	self.shards = shards
	for i := uint64(0); i < shards; i++ {
		self.bucket = append(self.bucket, NewSession(ttl, count, deferred, domains))
	}
	return
}

func (self * Sessions_t) Clear() {
	for _, b := range self.bucket {
		b.Clear()
	}
}

func (self * Sessions_t) Flush(LastTs int64, keep int, evicted Evict) {
	for _, b := range self.bucket {
		b.Flush(LastTs, keep, evicted)
	}
}

func (self * Sessions_t) Remove(ShardKey uint64, Domain interface{}, UID interface{}, evicted Evict) bool {
	return self.bucket[ShardKey % self.shards].Remove(Domain, UID, evicted)
}

func (self * Sessions_t) Update(ShardKey uint64, Ts int64, Domain interface{}, UID interface{}, NewData NewData_t, evicted Evict) (Diff int64, Mapped Mapped_t) {
	return self.bucket[ShardKey % self.shards].Update(Ts, Domain, UID, NewData, evicted)
}

func (self * Sessions_t) ListFront(evicted Evict) {
	for _, b := range self.bucket {
		if b.ListFront(evicted) == false {
			return
		}
	}
}

func (self * Sessions_t) ListBack(evicted Evict) {
	for _, b := range self.bucket {
		if b.ListBack(evicted) == false {
			return
		}
	}
}

func (self * Sessions_t) Stat(ShardKey uint64, Domain interface{}) (stat Stat_t) {
	return self.bucket[ShardKey % self.shards].Stat(Domain)
}

func (self * Sessions_t) StatBuckets() (res []StatList_t) {
	for _, b := range self.bucket {
		res = append(res, b.StatList()...)
	}
	return
}

func (self * Sessions_t) Size() (x int, y int) {
	for _, b := range self.bucket {
		_x, _y := b.Size()
		x += _x
		y += _y
	}
	return
}

func (self * Sessions_t) SizeBuckets() (res [][]int) {
	for _, b := range self.bucket {
		x, y := b.Size()
		res = append(res, []int{x, y})
	}
	return
}
