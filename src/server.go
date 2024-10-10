package main

import (
	"net"
	"log"
)

type Connections struct {
	alive []bool
	conn []net.Conn
	out map[int]chan []byte 
	in map[int]chan []byte 
}

type Server struct {
	connections Connections
	listener net.Listener
}

func handleRead(connections Connections, conn_id int) {
	buf := make([]byte, SERVER_BYTE_RECEIVE_SIZE, SERVER_BYTE_RECEIVE_SIZE)
	for true {
		read, err := connections.conn[conn_id].Read(buf)
		if err != nil {
			log.Printf("ERROR: conn_id = %d, read = %v, err = %s\n", conn_id, read, err)
			break
		}
		connections.in[conn_id] <- buf
		if !connections.alive[conn_id] {
			break
		}
	}
	if connections.alive[conn_id] {
		connections.alive[conn_id] = false
		connections.conn[conn_id].Close()
	}
}

func handleWrite(connections Connections, conn_id int) {
	buf := make([]byte, SERVER_BYTE_SEND_SIZE, SERVER_BYTE_SEND_SIZE)
	for true {
		buf = <- connections.out[conn_id]
		if !connections.alive[conn_id] {
			break
		}
		written, err := connections.conn[conn_id].Write(buf)
		if err != nil {
			log.Printf("ERROR: conn_id = %d, written = %v, err = %s\n", conn_id, written, err)
			break
		}
	}
	if connections.alive[conn_id] {
		connections.alive[conn_id] = false
		connections.conn[conn_id].Close()
	}
}

func server_init(server *Server, port string, player_num int) {
	log.Printf("listening on port %v\n", port)
	log.Printf("Waiting for %d player(s)\n", player_num - len(server.connections.conn))
	ln, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("Failed to listen to port %v: %s\n", port, err)
	}
	server.listener = ln
	id := 0

	server.connections.out = make(map[int]chan []byte)
	server.connections.in = make(map[int]chan []byte)

	for len(server.connections.conn) < player_num {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection %s", err)
			continue
		}
		addr := conn.RemoteAddr()
		log.Printf("%s connected\n", addr.String())
		log.Printf("Waiting for %d player(s)\n", player_num - len(server.connections.conn))
		server.connections.alive = append(server.connections.alive, true)
		server.connections.conn = append(server.connections.conn, conn)
		server.connections.out[id] = make(chan []byte)
		server.connections.in[id] = make(chan []byte)
		go handleRead(server.connections, id)
		go handleWrite(server.connections, id)
		id += 1
	}
}

func server_close(server *Server) {
	server.listener.Close()
}