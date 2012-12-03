package main

import (
	"net"
	"os"
	"fmt"
	"strings"
	"container/list"
	"github.com/thoj/go-mysqlpure"
	"bytes"
)

var waiting *list.List;
var games map[string] Game

//Holds information pertaining to a single player
type Person struct {
	name string
	con net.Conn
}

//Holds information pertaining to a single game between two players
type Game struct {
	people map[string] Person
}

func main() {
	waiting = list.New()
	games = make(map[string]Game)
	LogIt("SETUP", "Starting...")
	
	//Setup server socket
	addr, err := net.ResolveTCPAddr("ip4", ":4849")
	ErrorCheck(err, "Problem resolving TCP address")
	
	//Listen on socket
	listen, err := net.ListenTCP("tcp", addr)
	ErrorCheck(err, "TCP listening error")
	
	LogIt("SETUP", "Ready.")

	for{
		//Wait for connection
		connection, err := listen.Accept()
		if(err != nil){
			continue
		}
		LogIt("CONNECTION", "Got new connection")
		
		//Setup connection to new client in it's own thread
		go newClient(connection)
		
	}

	os.Exit(0)
}

//Called when a new client connects to the server.
//Reads command from the client and sends it to be parsed.
//It then sits awaiting more commands from the client
func newClient(connect net.Conn){
	LogIt("CONNECTION", "Handling new client")
	var buffer [512]byte

	_, err := connect.Read(buffer[0:])
	if err != nil {
		LogError("ERROR", "Error reading from client", err)
		connect.Close()
		return
	}

	parseCommand(string(bytes.TrimRight(buffer[0:], string(byte(0)))), connect)
	for {
		_, err := connect.Read(buffer[0:])
		if err != nil {
			LogError("ERROR", "Error reading from client", err)
			connect.Close()
			return
		}
		parseCommand(string(bytes.TrimRight(buffer[0:], string(byte(0)))), connect)
	}
}

//Parses out client commands
func parseCommand(com string, connection net.Conn){

	//var response string;
	parts := strings.Split(com, ":")
	dataCon, err := mysql.Connect("tcp", "127.0.0.1:3306", "hhss", "highscores", "hhss")
	ErrorCheck(err, "Could not connect to MySQL database.")

	switch parts[0] {
		//Creates a new game or adds user to a game waiting for a partner
		case "new":
			checker := new(mysql.MySQLResponse)

			checker, err = dataCon.Query("SELECT username FROM users WHERE username='" + parts[1] + "';")
			if len(checker.FetchRowMap()) > 0{
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
			} else {
				connection.Write([]byte("fail:you don't exist"))
				connection.Close();
			}
		//Passes along a player's move to another player
		case "move":
			fmt.Println(parts)
			fmt.Println("move:" + parts[1] + ":" + parts[3] + ":" + parts[4])
			 _, err := games[parts[2]].people[parts[1]].con.Write([]byte("move:" + parts[1] + ":" + parts[3] + ":" + parts[4]))
			if ErrorCheck(err, "Could not send new move to client in game " + parts[2]){
				connection.Write([]byte("fail:Could not message opponent."))
			}
		//Passess along a player winning to another player
		case "finished":
			_, err := games[parts[2]].people[parts[1]].con.Write([]byte("finished:" + parts[2] + ":" + parts[3] + ":" + parts[4]))
			if ErrorCheck(err, "Could not send finished message to client in game " + parts[2]) {
				connection.Write([]byte("fail:Could not message opponent."))
			}
			for _, p := range(games[parts[2]].people){
				p.con.Close();
			}
	}
	dataCon.Quit();
}

//Creates new game
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