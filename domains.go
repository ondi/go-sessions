//
//
//

package sessions

type Domains interface {
	Add(Domain interface{})
	Remove(Domain interface{}, Hits int64, Diff int64)
	Update(Domain interface{}, Hits int64, Diff int64)
	Clear()
	Stat(Domain interface{}) Stat_t
	StatList() (res []StatList_t)
	Size() int
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

func (self * Domains_t) Clear() {
	self.stats = map[interface{}]*Stat_t{}
}

func (self * Domains_t) Add(Domain interface{}) {
	if stat, ok := self.stats[Domain]; !ok {
		self.stats[Domain] = &Stat_t{Hits: 1, Sessions: 1, Bounces: 1, Duration: 0}
	} else {
		stat.Hits++
		stat.Sessions++
		stat.Bounces++
	}
}

func (self * Domains_t) Remove(Domain interface{}, Hits int64, Diff int64) {
	stat := self.stats[Domain]
	if stat.Sessions > 1 {
		stat.Sessions--
		if Hits == 1 {
			stat.Bounces--
		}
		stat.Hits -= Hits
		stat.Duration -= Diff
	} else {
		delete(self.stats, Domain)
	}
}

func (self * Domains_t) Update(Domain interface{}, Hits int64, Diff int64) {
	stat := self.stats[Domain]
	stat.Hits++
	if Hits == 2 {
		stat.Bounces--
	}
	stat.Duration += Diff
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

func (self * Domains_t) Size() int {
	return len(self.stats)
}

type NoDomains_t struct {}

func NewNoDomains() (* NoDomains_t) {
	return &NoDomains_t{}
}

func (* NoDomains_t) Add(Domain interface{}) {}
func (* NoDomains_t) Remove(Domain interface{}, Hits int64, Diff int64) {}
func (* NoDomains_t) Update(Domain interface{}, Hits int64, Diff int64) {}
func (* NoDomains_t) Clear() {}
func (* NoDomains_t) Stat(Domain interface{}) (res Stat_t) {return}
func (* NoDomains_t) StatList() (res []StatList_t) {return}
func (* NoDomains_t) Size() int {return 0}
