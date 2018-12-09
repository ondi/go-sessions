package sessions

import "testing"

func ExampleSort1() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, false, NewDomains())
	cc.Update(1, 1, 1, 1, NoNewData_t{}, &evicted)
	
/* Output:
*/
}

func ExampleSort2() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, false, NewNoDomains())
	cc.Update(1, 1, 1, 1, NoNewData_t{}, &evicted)
	
/* Output:
*/
}

func Test1(t * testing.T) {

}
