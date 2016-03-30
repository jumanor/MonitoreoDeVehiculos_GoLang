package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/googollee/go-socket.io"
)

// Función middleware agrega cabeceras CORS a cada solicitud
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

// Función buscar retorna -1 si el ID no se encontró
// en el listado de clientes o la posición en el slice
// si se encontró
func buscar(id string) int {
	var index = -1
	for i := 0; i < len(CLIENTES); i++ {
		if CLIENTES[i].id == id {
			return i
		}
	}
	return index
}

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("CONNECTED key:" + so.Id())

		so.On("loginMonitor", func(monitor string) {
			if monitor == "monitor" {
				so.Join("monitores")
				so.Emit("loginMonitorResponse", string(so.Id()))
			}
		})

		so.On("loginCliente", func(id string) {
			var index = buscar(id)
			if index == -1 {
				var tmp Cliente
				tmp.id = id
				tmp.socket_id = so.Id()

				CLIENTES = append(CLIENTES, &tmp)
				jsonData := map[string]string{"estado": "1", "id": id}
				var jsonStringify, _ = json.Marshal(jsonData)

				so.Emit("loginClienteResponse", string(jsonStringify))
			} else {
				jsonData := map[string]string{"estado": "0", "id": ""}
				var jsonStringify, _ = json.Marshal(jsonData)

				so.Emit("loginClienteResponse", string(jsonStringify))
			}
		})

		so.On("posicionClientes", func(data string) {
			log.Println(data)
			so.BroadcastTo("monitores", "monitorPrincipal", data)
		})

		so.On("disconnection", func() {
			log.Println("DISCONNECT key:" + so.Id())

			for i := 0; i < len(CLIENTES); i++ {
				if CLIENTES[i].socket_id == so.Id() {
					jsonData := map[string]string{"id": CLIENTES[i].id, "socket_id": CLIENTES[i].socket_id}
					var jsonStringify, _ = json.Marshal(jsonData)

					so.BroadcastTo("monitores", "monitorPrincipalDisconnet", string(jsonStringify))

					CLIENTES = append(CLIENTES[:i], CLIENTES[i+1:]...)
				}
			}

		})
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", middleware(server))
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	log.Println("Serving at localhost:9095...")
	log.Fatal(http.ListenAndServe("127.0.0.1:9095", nil))
}
