//
//
//

package sessions

type Domains interface {
	NewUID(Domain interface{}, Data Data_t)
	Update(Domain interface{}, Hits int64, Duration int64)
	Remove(Domain interface{}, Hits int64, Duration int64)
	Size() int
	Clear()
}

type Stats interface {
	Stat(Domain interface{}) Stat_t
	StatList() (res []StatList_t)
}

type Stat_t struct {
	Hits int64
	Sessions int64
	Bounces int64
	Duration int64
}

type StatList_t struct {
	Domain interface{}
	Stat Stat_t
}

type Domains_t struct {
	stats map[interface{}]*Stat_t
}

func NewDomains() (* Domains_t) {
	return &Domains_t{stats: map[interface{}]*Stat_t{}}
}

func (self * Domains_t) NewUID(Domain interface{}, Data Data_t) {
	if stat, ok := self.stats[Domain]; !ok {
		self.stats[Domain] = &Stat_t{Hits: 1, Sessions: 1, Bounces: 1, Duration: 0}
	} else {
		stat.Hits++
		stat.Sessions++
		stat.Bounces++
	}
}

func (self * Domains_t) Update(Domain interface{}, Hits int64, Duration int64) {
	if stat, ok := self.stats[Domain]; ok {
		stat.Hits++
		if Hits == 2 {
			stat.Bounces--
		}
		stat.Duration += Duration
	}
}

func (self * Domains_t) Remove(Domain interface{}, Hits int64, Duration int64) {
	if stat, ok := self.stats[Domain]; !ok {
		return
	} else if stat.Sessions > 1 {
		stat.Sessions--
		if Hits == 1 {
			stat.Bounces--
		}
		stat.Hits -= Hits
		stat.Duration -= Duration
	} else {
		delete(self.stats, Domain)
	}
}

func (self * Domains_t) Size() int {
	return len(self.stats)
}

func (self * Domains_t) Clear() {
	self.stats = map[interface{}]*Stat_t{}
}

func (self * Domains_t) Stat(Domain interface{}) Stat_t {
	if res, ok := self.stats[Domain]; ok {
		return *res
	}
	return Stat_t{}
}

func (self * Domains_t) StatList() (res []StatList_t) {
	for k, v := range self.stats {
		res = append(res, StatList_t{k, *v})
	}
	return
}

type NoDomains_t struct {}

func NewNoDomains() (* NoDomains_t) {
	return &NoDomains_t{}
}

func (* NoDomains_t) NewUID(Domain interface{}, Data Data_t) {}
func (* NoDomains_t) Update(Domain interface{}, Hits int64, Duration int64) {}
func (* NoDomains_t) Remove(Domain interface{}, Hits int64, Duration int64) {}
func (* NoDomains_t) Size() int {return 0}
func (* NoDomains_t) Clear() {}
