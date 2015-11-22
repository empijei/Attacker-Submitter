package main
import (
	"net"
)
func send(flags string) {
	conn, err := net.Dial("tcp", "localhost:31337")
	fmt.Fprintf(conn, flags)
	conn.Close()
}
func main(){
	flags:=`FLAG1
	FLAG2
	FLAG`	
	send(flags)
}

