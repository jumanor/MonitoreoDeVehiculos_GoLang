package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/googollee/go-socket.io"
)

func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if origin := req.Header.Get("Origin"); origin != "" {
			rw.Header().Set("Access-Control-Allow-Origin", origin)
		}

		rw.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		rw.Header().Set("Access-Control-Allow-Credentials", "true")

		h.ServeHTTP(rw, req)
	})
}

// Estructura cliente maneja los IDs de cliente y
// de socket de un cliente conectado
type Cliente struct {
	id        string
	socket_id string
}

// Variable CLIENTES contiene un listado de todos
// los clientes encontrados
var CLIENTES []*Cliente

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("CONNECTED key:" + so.Id())
		so.On("loginMonitor", loginMonitor(so))
		so.On("loginCliente", loginCliente(so))
		so.On("posicionClientes", posicionClientes(so))
		so.On("disconnection", disconnection(so))
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", middleware(server))
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	log.Println("Serving at localhost:9095...")
	log.Fatal(http.ListenAndServe("127.0.0.1:9095", nil))
}

func buscarEnSlice(id string) bool {
	for _, val := range CLIENTES {
		if val.id == id {
			return true
		}
	}

	return false
}

func loginMonitor(so socketio.Socket) func(monitor string) {
	return func(monitor string) {
		if monitor != "monitor" {
			return
		}

		so.Join("monitores")
		so.Emit("loginMonitorResponse", string(so.Id()))
	}
}

func loginCliente(so socketio.Socket) func(id string) {
	return func(id string) {
		if buscarEnSlice(id) == false {
			tmp := Cliente{id: id, socket_id: so.Id()}
			CLIENTES = append(CLIENTES, &tmp)

			jsonData := map[string]string{
				"estado": "1",
				"id":     id,
			}

			barray, _ := json.Marshal(jsonData)
			so.Emit("loginClienteResponse", string(barray))
			return
		}

		jsonData := map[string]string{
			"estado": "0",
			"id":     "",
		}

		barray, _ := json.Marshal(jsonData)
		so.Emit("loginClienteResponse", string(barray))
	}
}

func posicionClientes(so socketio.Socket) func(data string) {
	return func(data string) {
		log.Println(data)
		so.BroadcastTo("monitores", "monitorPrincipal", data)
	}
}

func disconnection(so socketio.Socket) func() {
	return func() {
		log.Println("DISCONNECT key:" + so.Id())

		for i, cliente := range CLIENTES {
			if cliente.socket_id == so.Id() {
				jsonData := map[string]string{
					"id":        cliente.id,
					"socket_id": cliente.socket_id,
				}

				barray, _ := json.Marshal(jsonData)
				so.BroadcastTo("monitores", "monitorPrincipalDisconnet", string(barray))

				CLIENTES = append(CLIENTES[:i], CLIENTES[i+1:]...)
			}
		}
	}
}
