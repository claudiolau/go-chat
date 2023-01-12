package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct{ 
	conns map[*websocket.Conn]bool
}

// client sub to websocket handler
func NewServer() *Server{

	// maps in golang is not concurrent safe
	return &Server{ 
		conns:make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handlerWs(ws *websocket.Conn){
	fmt.Println("New incoming connection from client: ", ws.RemoteAddr())

	// optional use mutex to make sure do not have any race condition
	s.conns[ws]=true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn){ 
	buf := make([]byte, 1024)
	for { 
		n, err := ws.Read(buf)
		if err != nil{ 
			if (err == io.EOF){ 
				break
			}
			fmt.Println("read error:", err)
			continue
		}
		msg := buf[:n]
		s.broadcast(msg)

	}
	
}

func (s *Server) handleWSOrderbook(ws *websocket.Conn){ 
	fmt.Println("New incoming connection from client to orderbook feed:", ws.RemoteAddr())

	for { 
		payload := fmt.Sprintf("orderbook data -> %d\n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(time.Second * 2)
	}
}

func (s *Server) broadcast(b []byte){ 
	for ws := range s.conns{ 
		go func(ws *websocket.Conn) { 
			if _, err := ws.Write(b); err != nil{
				fmt.Println("write error:", err)
			}
			
		}(ws)
	}
}

func main(){ 

	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handlerWs))
	http.Handle("/orderbookfeed", websocket.Handler(server.handleWSOrderbook))
	http.ListenAndServe(":3000", nil)

}