//
// Session Tracking Shards
//

package sessions

type ID64_t interface {
	Sum64() uint64
}

type Sessions_t struct {
	shards uint64
	bucket []*Session_t
}

func NewSessions(shards uint64, ttl int64, count int) (self * Sessions_t) {
	self = &Sessions_t{}
	self.shards = shards
	for i := uint64(0); i < shards; i++ {
		self.bucket = append(self.bucket, NewSession(ttl, count))
	}
	return
}

func (self * Sessions_t) get_bucket(Domain ID64_t) uint64 {
	return Domain.Sum64() % self.shards
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

func (self * Sessions_t) Remove(Domain ID64_t, UID interface{}, evicted Evict) bool {
	i := self.get_bucket(Domain)
	return self.bucket[i].Remove(Domain, UID, evicted)
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

func (self * Sessions_t) Update(Ts int64, Domain ID64_t, UID interface{}, Data func () interface{}, evicted Evict) (Diff int64, Mapped Mapped_t) {
	i := self.get_bucket(Domain)
	return self.bucket[i].Update(Ts, Domain, UID, Data, evicted)
}

func (self * Sessions_t) Stat(Domain ID64_t) (stat Stat_t) {
	i := self.get_bucket(Domain)
	return self.bucket[i].Stat(Domain)
}

func (self * Sessions_t) StatList() (res []StatRow_t) {
	for _, b := range self.bucket {
		res = append(res, b.StatList()...)
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

func (self * Sessions_t) Size() (res [][]int) {
	temp := []int{0, 0, 0}
	for _, b := range self.bucket {
		x, y := b.Size()
		temp[1] += x
		temp[2] += y
	}
	temp[0] = int(self.shards)
	return append(res, temp)
}
