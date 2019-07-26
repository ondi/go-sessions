package sessions

import "testing"

func ExampleSort1() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, NewDomains())
	cc.Update(1, 1, 1, 1, func() interface{}{return nil}, &evicted)
	
/* Output:
*/
}

func ExampleSort2() {
	var evicted Drop_t
	cc := NewSessions(1, 15, 10, nil)
	cc.Update(1, 1, 1, 1, func() interface{}{return nil}, &evicted)
	
/* Output:
*/
}

func Test1(t * testing.T) {

}
