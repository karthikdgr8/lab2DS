package becauseGO

import "net"

type Callback struct {
	Callback func(conn net.Conn)
}

func (a Callback) SetCallback(myfunc func(conn net.Conn)) {
	a.Callback = myfunc
}
