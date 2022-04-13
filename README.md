# Golang Server for Gamemaker Studio games
This is a simple TCP server that can handle buffers coming from games made in Gamemaker Studio. For the Gamemaker client example, visit here: https://github.com/LDinos/gms_go_client

# Running
Use ```go run .```

# Documentation
Everytime a client connects, we save them as a key to the map called `clients`. The value for each key is unused (I am adding 0 by default) but you can use it to store usernames etc. for each client if you desire. Obviously, when they disconnect, their map key is deleted.
### SEND BUFFERS TO CLIENT:
Use `binary.Write(conn, binary.LittleEndian, var)`. Usually you send one byte that expresses what to expect
afterwards. We call that byte the "Buffer Message". Buffer Messages are written as Macros in Gamemaker (buffer_messages.go in here)
Buffer Messages are unsigned 8 bits, which means they can reach up to 255. Feel free to change it to u16
if you need more. Buffer Messages that express buffers that THE SERVER will send to the client have the
keyword _SEND_ in them. Example: Suppose the client asks us to send the number `19238`. We are first gonna let
the client know that the buffer we are gonna send them is about that number, so we first write
`binary.Write(conn, binary.LittleEndian, NET_SEND_REQUESTED_NUM)` and THEN we send the actual number
`binary.Write(conn, binary.LittleEndian, uint16(19238))`. Make sure Gamemaker has the same value for buffer
messages as in here. Eg ```NET_SEND_REQUESTED_NUM``` equals 2 here, so when gamemaker client gets the data, you
should expect to handle it with a macro like "`#macro NET_GET_REQUESTED_NUM 2`"

### RECEIVE BUFFERS FROM CLIENT:
When a Gamemaker game sends buffers, we handle them in the `handleConnection` function.
`buf[0]` expresses the Buffer Message, so we make a switch statement and then read the following
buffer(s) and their known types depending on the Buffer Message. Buffer Message buffers that we
receive have the keyword _GET_ in them. If the buffer you receive is a string, use `buffer_get_string`,
because each string character takes one slot in the buf array. The function will take care to stitch
each character together (eg: if buf[1] = "h", [2] = "e", [3] = "l", [4] = "l", [5] = "o", then
buffer_get_string will automatically return "hello"). Otherwise just use `buffer_get_number` or read from
the array 'buf' yourself (make sure to add +1 to the index if you use the last practice).

### BUFFER TYPES:
Make sure to use the functions needed when SENDING a value for the client. Gamemaker must know
what types of variables to expect. These functions are `int8()`, `int16()`...etc for signed integers or
`uint8()`, `uint16()`... etc for unsigned. Use `[]byte(string_here)` to send strings.
Example: binary.Write(conn, binary.LittleEndian, uint16(89812))

# Communication Example
## From Client to Server
**Gamemaker:** 
```
buffer_seek(c_buffer, buffer_seek_start, 0)
buffer_write(c_buffer, buffer_u8, NET_SEND_TESTMSG) //buffer message
buffer_write(c_buffer, buffer_string, "TEST123") //a string
buffer_write(c_buffer, buffer_u8, 20) //a number
network_send_raw(client, c_buffer, buffer_tell(c_buffer)) //Ignore the warning
```

Gamemaker client just sent this buffer to the server with a buffer message, a string and a number. Let's see how our server handles that:

**Go:**
```
...
index := 1
	switch buf[0] {
	case NET_GET_TESTMSG:
		fmt.Println("I got this message:", buffer_get_string(buf, &index, reqLen))
		fmt.Println("I got this number:", buffer_get_number(buf, &index))
	...
```
Go server understands and reads both the string and the number and shows it on console.

## From Server to Client
**Go:**
```
binary.Write(conn, binary.LittleEndian, NET_SEND_REQUESTED_NUM)
binary.Write(conn, binary.LittleEndian, uint16(19238)) //number that the server will send
```
Go server sends a number to one client (conn).

**Gamemaker:**
```
...
switch(cmd_type) {
		case NET_GET_REQUESTED_NUM:
			var s = buffer_read(t_buffer, buffer_u16);
			num = string(s)
			show_debug_message("this is the number I got: " + num)
			break;
```
Gamemaker succesfully gets the number and shows it.

## From Server to multiple Clients
If you want to send a buffer to all connected clients of the server, you can write something like this:
```
for client := range clients { //send to all connected clients
			binary.Write(client, binary.LittleEndian, NET_BROADCAST_MSG)
			binary.Write(client, binary.LittleEndian, []byte(msg)) //convert string to byte
		}
```