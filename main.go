package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

//Estructura para chat
type Chat struct {
	mu        sync.Mutex
	usuarios  map[string]*Usuario
	broadcast chan []byte
}
type Usuario struct{
	ID string 
	Conn *websocket.Conn
}
//Estructura para multiples chats
var chats = make(map[string]*Chat)
var chatsMutex = &sync.Mutex{}
//Upgrader para http to ws
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	//Ruta de upgrader
	http.HandleFunc("/ws", wsHandler)
	//Servir HTML en /
	//Escuchando en puerto 9012
	http.ListenAndServe(":9012", nil)
}

func getOrCreateChat(chatID string) *Chat {
	//Recibe chatID y busca dentro de el array de chats
	chatsMutex.Lock()
	temp, found := chats[chatID]
	if found {
		chatsMutex.Unlock()
		return temp
	}
	//Crea un valor nuevo de Chat dentro de la variable chat
	chat := &Chat{
		usuarios:   make(map[string]*Usuario),
		broadcast: make(chan []byte),
	}
	//Asigna valor a mapa ([K]String -> [V]Chat)
	chats[chatID] = chat
	chatsMutex.Unlock()

	/*
		Función "while true" q se encarga de repartir el mensaje dentro de un
		map de usuarios
	*/
	go func() {
		for msg := range chat.broadcast {
			chat.mu.Lock()
			usuarios := make([]*Usuario, 0, len(chat.usuarios))
			for _, u := range chat.usuarios {
				usuarios = append(usuarios, u)
			}
			chat.mu.Unlock()
			
			for _, u := range usuarios {
				u.Conn.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}()
	
	return chat
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//Extrae la query de la url
	query := r.URL.Query()
	//Extrae el valor que se pasa por medio de el valor chat en la url
	chaId := query.Get("chat")

	//Chat id vacio
	if chaId == "" {
		http.Error(w, "Missing chat ID", http.StatusBadRequest)
		return
	}
	//Crear chat en el caso de que tenga un Id
	chat := getOrCreateChat(chaId)
	//Hace el upgrade de http a ws
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to websocket:", err)
		return
	}
	defer conn.Close()
	//Crea usuario nuevo del websocket
	usuario := &Usuario{
		ID:   uuid.NewString(),
		Conn: conn,
	}
	//Introduce el usuario a el array que tiene chat de usuarios
	chat.mu.Lock()
	chat.usuarios[usuario.ID] = usuario
	chat.mu.Unlock()
	// Se encarga de enviar todos los mensajes nuevos que lleguen a el broadcast del chat
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			chat.mu.Lock()
			delete(chat.usuarios, usuario.ID)
			chat.mu.Unlock()
			break
		}
		chat.broadcast <- msg
	}
}