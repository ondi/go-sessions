//
// Session Tracking Shards
//

package sessions

type Sessions_t struct {
	shards uint64
	bucket []*Session_t
}

func NewSessions(shards uint64, ttl int64, limit int, domains Domains, evict Evict) (self * Sessions_t) {
	self = &Sessions_t{}
	self.shards = shards
	for i := uint64(0); i < shards; i++ {
		self.bucket = append(self.bucket, NewSession(ttl, limit, domains, evict))
	}
	return
}

func (self * Sessions_t) Clear() {
	for _, b := range self.bucket {
		b.Clear()
	}
}

func (self * Sessions_t) Flush(Ts int64, keep int) {
	for _, b := range self.bucket {
		b.Flush(Ts, keep)
	}
}

func (self * Sessions_t) Remove(ShardKey uint64, Domain interface{}, UID interface{}) bool {
	return self.bucket[ShardKey % self.shards].Remove(Domain, UID)
}

func (self * Sessions_t) Update(ShardKey uint64, Ts int64, Domain interface{}, UID interface{}, NewData func() interface{}) (Diff int64, Mapped Mapped_t) {
	return self.bucket[ShardKey % self.shards].Update(Ts, Domain, UID, NewData)
}

func (self * Sessions_t) ListFront(list Evict) {
	for _, b := range self.bucket {
		if b.ListFront(list) == false {
			return
		}
	}
}

func (self * Sessions_t) ListBack(list Evict) {
	for _, b := range self.bucket {
		if b.ListBack(list) == false {
			return
		}
	}
}

func (self * Sessions_t) Stat(ShardKey uint64, Domain interface{}) (stat Stat_t) {
	return self.bucket[ShardKey % self.shards].Stat(Domain)
}

func (self * Sessions_t) StatBuckets() (res map[interface{}]Stat_t) {
	for _, b := range self.bucket {
		for k, v := range b.StatList() {
			res[k] = v
		}
	}
	return
}

func (self * Sessions_t) Size() (res int) {
	for _, b := range self.bucket {
		res += b.Size()
	}
	return
}

func (self * Sessions_t) DomainSize() (res int) {
	for _, b := range self.bucket {
		res += b.DomainSize()
	}
	return
}

func (self * Sessions_t) SizeBuckets() (res []int) {
	for _, b := range self.bucket {
		res = append(res, b.Size())
	}
	return
}

func (self * Sessions_t) DomainSizeBuckets() (res []int) {
	for _, b := range self.bucket {
		res = append(res, b.DomainSize())
	}
	return
}
