package sessions

import "testing"

type MyData_t struct {}

func (self * MyData_t) Lock() {}

func (self * MyData_t) NewData() Data_t {
	return &MyData_t{}
}

func ExampleSort1() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, false, NewDomains())
	cc.Update(1, 1, 1, 1, (* MyData_t)(nil), &evicted)
	
/* Output:
*/
}

func ExampleSort2() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, false, NewNoDomains())
	cc.Update(1, 1, 1, 1, (* MyData_t)(nil), &evicted)
	
/* Output:
*/
}

func Test1(t * testing.T) {

}
