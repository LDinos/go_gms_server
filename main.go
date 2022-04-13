package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

const max_clients = 8

var clients = make(map[net.Conn]int)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
		fmt.Print(err)
	} else {
		fmt.Println("Server is now open.")
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Print(err)
		}
		fmt.Println("New client:", conn.LocalAddr().String())
		if add_client(conn) {
			go handleConnection(conn)
		} else {
			binary.Write(conn, binary.LittleEndian, NET_SEND_KICK)
			binary.Write(conn, binary.LittleEndian, []byte("Full Server"))
			fmt.Println("Client tried to connect on full server", conn.LocalAddr().String())
		}
	}

}

/* If we get new data from a client, we run this */
func handleConnection(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			remove_client(conn)
			return
		}
		handleMessage(conn, buf, reqLen)
	}
}

/* Handle received buffers here */
func handleMessage(conn net.Conn, buf []byte, reqLen int) {
	index := 1
	switch buf[0] {
	case NET_GET_TESTMSG:
		fmt.Println("I got this message:", buffer_get_string(buf, &index, reqLen))
		fmt.Println("I got this number:", buffer_get_number(buf, &index))
	case NET_GET_REQUESTNUM:
		binary.Write(conn, binary.LittleEndian, NET_SEND_REQUESTED_NUM)
		binary.Write(conn, binary.LittleEndian, uint16(19238)) //random number I will send
	case NET_GET_MSG:
		msg := buffer_get_string(buf, &index, reqLen)
		fmt.Println("I will send this to everyone now:", msg)
		for client := range clients { //send to all connected clients
			binary.Write(client, binary.LittleEndian, NET_BROADCAST_MSG)
			binary.Write(client, binary.LittleEndian, []byte(msg))
		}
	}
}

func add_client(conn net.Conn) bool {
	if len(clients) >= max_clients {
		return false
	}
	clients[conn] = 0 //You can store whatever you want here, like usernames etc. Im just adding 0 since I dont have a use
	return true
}

func remove_client(conn net.Conn) {
	delete(clients, conn)
}

/* Use this when expecting a number buffer */
func buffer_get_number(buff []byte, index *int) byte {
	val := buff[*index]
	*index += 1
	return val
}

/* Use this when expecting a string buffer */
func buffer_get_string(buff []byte, index *int, l int) string {
	str := ""
	start := *index
	for i := start; i < l; i++ {
		*index = i
		if buff[i] == 0 {
			*index += 1
			break
		} else {
			str += string(buff[i])
		}
	}
	return str
}

/*
	SEND BUFFERS TO CLIENT:
Use binary.Write(conn, binary.LittleEndian, var). Usually you send one byte that expresses what to expect
after. We call that byte the "Buffer Message". Buffer Messages are written as Macros in Gamemaker (buffer_messages.go in here)
Buffer Messages are unsigned 8 bits, which means they can reach up to 255. Feel free to change it to u16
if you need more. Buffer Messages that express buffers that THE SERVER will send to the client have the
keyword SEND in them. Example: Suppose the client asks us to send the number 19238. We are first gonna let
the client know that the buffer we are gonna send them is about that number, so we first write
binary.Write(conn, binary.LittleEndian, NET_SEND_REQUESTED_NUM) and THEN we send the actual number
binary.Write(conn, binary.LittleEndian, uint16(19238)). Make sure Gamemaker has the same value for buffer
messages as in here. Eg NET_SEND_REQUESTED_NUM equals 2 here, so when gamemaker client gets the data, you
should expect to handle it with a macro like "#macro NET_GET_REQUESTED_NUM 2"

	RECEIVE BUFFERS FROM CLIENT:
When a Gamemaker game sends buffers, we handle them in the handleConnection function.
buf[0] expresses the Buffer Message, so we make a switch statement and then read the following
buffer(s) and their known types depending on the Buffer Message. Buffer Message buffers that we
receive have the keyword GET in them. If the buffer you receive is a string, use buffer_get_string,
because each string character takes one slot in the buf array. The function will take care to stitch
each character together (eg: if buf[1] = "h", [2] = "e", [3] = "l", [4] = "l", [5] = "o", then
buffer_get_string will automatically return "hello"). Otherwise just use buffer_get_number or read from
the array 'buf' yourself (make sure to add +1 to the index if you use the last practice).

	BUFFER TYPES:
Make sure to use the functions needed when SENDING a value for the client. Gamemaker must know
what types of variables to expect. These functions are int8(), int16...etc for signed integers or
uint8(), uint16()... etc for unsigned. string() is used for strings.
Example: binary.Write(conn, binary.LittleEndian, uint16(89812))
*/
