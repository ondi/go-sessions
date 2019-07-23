package sessions

import "testing"

func ExampleSort1() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, NewDomains(), nil)
	cc.Update(1, 1, 1, 1, &evicted)
	
/* Output:
*/
}

func ExampleSort2() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, nil, nil)
	cc.Update(1, 1, 1, 1, &evicted)
	
/* Output:
*/
}

func Test1(t * testing.T) {

}
