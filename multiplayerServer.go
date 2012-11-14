package main

import (
	"net"
	"os"
	"fmt"
	"strings"
	"container/list"
	//"time"
)

var waiting *list.List;
var games map[string] Game


type Person struct {
	name string
	con net.Conn
}

type Game struct {
	people map[string] Person
}

func main() {
	waiting = list.New()
	games = make(map[string]Game)
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

	_ = parseCommand(string(buffer[0:]), connect)
}


func parseCommand(com string, connection net.Conn){

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
		case "move":
			 _, err := games[parts[2]].people[parts[1]].con.Write([]byte("move:" + parts[1] + ":" + parts[3] + ":" + parts[4]))
			if ErrorCheck(err, "Could not send new move to client in game " + parts[2]){
				connection.Write([]byte("fail:Could not message opponent."))
			}
			//fmt.Println("%s: MOVED in game %s: %s, %s", parts[1], parts[2], parts[3], parts[4])
		case "finished":
			_, err := games[parts[2]].people[parts[1]].con.Write([]byte("finished:" + parts[2] + ":" + parts[3] + ":" + parts[4]))
			if ErrorCheck(err, "Could not send finished message to client in game " + parts[2]) {
				connection.Write([]byte("fail:Could not message opponent."))
			}
			for _, p := range(games[parts[2]].people){
				p.con.Close();
			}
	}
}

func newGame(p1 Person, p2 Person) {
	gameName := p1.name + "AND" + p2.name
	fmt.Println(gameName)

	games[gameName] = Game{make(map[string]Person)}
	games[gameName].people[p1.name] = p2
	games[gameName].people[p2.name] = p1
	
	p1.con.Write([]byte("partner:" + p2.name + ":" + gameName))
	p2.con.Write([]byte("partner:" + p1.name + ":" + gameName))
}

//LogIt("CONNECTION", "Closing connection to clients " + p1.name + " and " + p2.name)