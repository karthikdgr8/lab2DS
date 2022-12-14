package becauseGO

import "net"

type Callback struct {
	Callback func(data []byte, conn net.Conn)
}

func (a Callback) SetCallback(myfunc func(data []byte, conn net.Conn)) {
	a.Callback = myfunc
}
