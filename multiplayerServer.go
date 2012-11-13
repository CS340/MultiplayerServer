package main

import (
	"net"
	"os"
	"fmt"
	"strings"
	"container/list"
	"time"
	"./game"
	"./logging"
)

var waiting *list.List;


type Person struct {
	name string;
	con net.Conn;
}

func main() {
	waiting = list.New()
	LogIt("SETUP", "Starting...")
	
	addr, err := net.ResolveTCPAddr("ip4", ":4849")
	ErrorCheck(err, "Problem resolving TCP address")
	
	listen, err := net.ListenTCP("tcp", addr)
	ErrorCheck(err, "TCP listening error")
	
	LogIt("SETUP", "Ready.")

	for{
		connection, err := listen.Accept()
		if(err != nil){
			continue
		}
		LogIt("CONNECTION", "Got new connection")
		
		go newClient(connection)
		
	}

	os.Exit(0)
}

func newClient(connect net.Conn){
	LogIt("CONNECTION", "Handling new client")
	var buffer [512]byte

	_, err := connect.Read(buffer[0:])
	if err != nil {
		LogError("ERROR", "Error reading from client", err)
		connect.Close()
		return
	}

	commm := parseCommand(string(buffer[0:]), connect)
	fmt.Println(commm)
	//_, err2 := connect.Write([]byte(commm))
	//if err2 != nil {
		//LogError("ERROR", "Error writing to client", err2)
		//connect.Close()
		//return
	//}
	//connect.Close()
	//LogIt("CONNECTION", "Closing connection to client")
}


func parseCommand(com string, connection net.Conn) (string){

	//var response string;
	parts := strings.Split(com, ":")

	switch parts[0] {
		case "new":
			var newPerson Person
			newPerson.name = parts[1]
			newPerson.con = connection

			waiting.PushFront(newPerson)

			if waiting.Len() > 1 {
				var p1,p2 Person
				e1 := waiting.Back()
				p1 = e1.Value.(Person)
				waiting.Remove(e1)
				e2 := waiting.Back()
				p2 = e2.Value.(Person)
				waiting.Remove(e2)
				go newGame(p1,p2)
			}
	}

	return "RESPONSE"
}

func newGame(p1 Person, p2 Person) {
	p1.con.Write([]byte(p2.name))
	p2.con.Write([]byte(p1.name))
	p1.con.Close()
	p2.con.Close()
	LogIt("CONNECTION", "Closing connection to clients " + p1.name + " and " + p2.name)
}

