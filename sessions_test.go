package sessions

import "testing"

func ExampleSort1() {
	cc := NewSessions(1, 15, 10, NewDomains(), Drop)
	cc.Update(1, 1, 1, 1, func() interface{}{return nil})
	
/* Output:
*/
}

func ExampleSort2() {
	cc := NewSessions(1, 15, 10, nil, Drop)
	cc.Update(1, 1, 1, 1, func() interface{}{return nil})
	
/* Output:
*/
}

func Test1(t * testing.T) {

}
