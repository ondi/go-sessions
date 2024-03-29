//
//
//

package sessions

type Domains interface {
	AddSession(Domain interface{}, Data interface{})
	RemoveSession(Domain interface{}, Hits int64, Duration int64)
	UpdateSession(Domain interface{}, Hits int64, Duration int64)
	Clear()
}

type Stats interface {
	Size() int
	Stat(Domain interface{}) Stat_t
	StatList() (res map[interface{}]Stat_t)
}

type Stat_t struct {
	Hits     int64
	Sessions int64
	Bounces  int64
	Duration int64
}

type Domains_t struct {
	stats map[interface{}]*Stat_t
}

func NewDomains() *Domains_t {
	return &Domains_t{stats: map[interface{}]*Stat_t{}}
}

func (self *Domains_t) AddSession(Domain interface{}, Data interface{}) {
	if stat, ok := self.stats[Domain]; !ok {
		self.stats[Domain] = &Stat_t{Hits: 1, Sessions: 1, Bounces: 1, Duration: 0}
	} else {
		stat.Hits++
		stat.Sessions++
		stat.Bounces++
	}
}

func (self *Domains_t) RemoveSession(Domain interface{}, Hits int64, Duration int64) {
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

func (self *Domains_t) UpdateSession(Domain interface{}, Hits int64, Duration int64) {
	if stat, ok := self.stats[Domain]; ok {
		stat.Hits++
		if Hits == 2 {
			stat.Bounces--
		}
		stat.Duration += Duration
	}
}

func (self *Domains_t) Clear() {
	self.stats = map[interface{}]*Stat_t{}
}

func (self *Domains_t) Size() int {
	return len(self.stats)
}

func (self *Domains_t) Stat(Domain interface{}) (res Stat_t) {
	if temp, ok := self.stats[Domain]; ok {
		res = *temp
	}
	return
}

func (self *Domains_t) StatList(out map[interface{}]Stat_t) {
	for k, v := range self.stats {
		out[k] = *v
	}
}

type NoDomains_t struct{}

func (NoDomains_t) AddSession(Domain interface{}, Data interface{})              {}
func (NoDomains_t) RemoveSession(Domain interface{}, Hits int64, Duration int64) {}
func (NoDomains_t) UpdateSession(Domain interface{}, Hits int64, Duration int64) {}
func (NoDomains_t) Size() int                                                    { return 0 }
func (NoDomains_t) Clear()                                                       {}
