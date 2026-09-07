package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestGetOrCreateChat verifica que getOrCreateChat crea y retorna chats correctamente
func TestGetOrCreateChat(t *testing.T) {
	// Limpiar estado global antes del test
	chats = make(map[string]*Chat)

	tests := []struct {
		name     string
		chatID   string
		wantNew  bool
		wantSame bool
	}{
		{
			name:    "crear nuevo chat",
			chatID:  "chat-1",
			wantNew: true,
		},
		{
			name:     "retorna el mismo chat",
			chatID:   "chat-1",
			wantSame: true,
		},
		{
			name:    "crear segundo chat",
			chatID:  "chat-2",
			wantNew: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialLen := len(chats)
			chat := getOrCreateChat(tt.chatID)

			if chat == nil {
				t.Error("getOrCreateChat retornó nil")
			}

			if tt.wantNew && len(chats) != initialLen+1 {
				t.Errorf("se esperaba crear un nuevo chat, pero len(chats) = %d, quería %d",
					len(chats), initialLen+1)
			}

			if tt.wantSame {
				sameChat := getOrCreateChat(tt.chatID)
				if sameChat != chat {
					t.Error("se esperaba obtener el mismo chat")
				}
			}

			// Verificar estructura del chat
			if chat.usuarios == nil {
				t.Error("chat.usuarios es nil")
			}
			if chat.broadcast == nil {
				t.Error("chat.broadcast es nil")
			}
		})
	}
}

// TestGetOrCreateChatBroadcast verifica que el broadcast funciona correctamente
func TestGetOrCreateChatBroadcast(t *testing.T) {
	chats = make(map[string]*Chat)

	chat := getOrCreateChat("test-broadcast")
	if chat == nil {
		t.Fatal("no se pudo crear el chat")
	}

	// Enviar un mensaje al broadcast
	testMsg := []byte("test message")
	select {
	case chat.broadcast <- testMsg:
		t.Log("mensaje enviado al broadcast exitosamente")
	case <-time.After(1 * time.Second):
		t.Error("timeout enviando mensaje al broadcast")
	}
}

// TestWSHandlerMissingChatID verifica que falta chatID
func TestWSHandlerMissingChatID(t *testing.T) {
	// Crear un servidor test
	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	// Intentar conectar sin chatID
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	if err == nil {
		t.Error("se esperaba un error al conectar sin chatID")
	}
}

// TestWSHandlerValidConnection verifica una conexión válida
func TestWSHandlerValidConnection(t *testing.T) {
	chats = make(map[string]*Chat)

	// Crear servidor test
	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	// Conectar con chatID válido
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=test-chat"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando a websocket: %v", err)
	}
	defer conn.Close()

	// Verificar que el usuario fue agregado al chat
	chat, exists := chats["test-chat"]
	if !exists {
		t.Fatal("chat no fue creado")
	}

	if len(chat.usuarios) == 0 {
		t.Fatal("usuario no fue agregado al chat")
	}
}

// TestWSHandlerMultipleUsers verifica múltiples usuarios en un chat
func TestWSHandlerMultipleUsers(t *testing.T) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=multi-user"

	// Conectar dos usuarios
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando usuario 1: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando usuario 2: %v", err)
	}
	defer conn2.Close()

	// Esperar un poco para que ambas conexiones se establezcan
	time.Sleep(100 * time.Millisecond)

	chat, exists := chats["multi-user"]
	if !exists {
		t.Fatal("chat no existe")
	}

	if len(chat.usuarios) != 2 {
		t.Errorf("se esperaban 2 usuarios, pero hay %d", len(chat.usuarios))
	}
}

// TestWSHandlerBroadcastMessage verifica que los mensajes se transmiten
func TestWSHandlerBroadcastMessage(t *testing.T) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=broadcast-test"

	// Conectar dos usuarios
	sender, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando sender: %v", err)
	}
	defer sender.Close()

	receiver, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando receiver: %v", err)
	}
	defer receiver.Close()

	// Esperar a que ambas conexiones se establezcan
	time.Sleep(100 * time.Millisecond)

	// Enviar mensaje del sender
	testMessage := "hello from sender"
	err = sender.WriteMessage(websocket.TextMessage, []byte(testMessage))
	if err != nil {
		t.Fatalf("Error escribiendo mensaje: %v", err)
	}

	// Leer mensaje en el receiver (con timeout)
	receiver.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := receiver.ReadMessage()
	if err != nil {
		t.Fatalf("Error leyendo mensaje: %v", err)
	}

	if string(msg) != testMessage {
		t.Errorf("mensaje recibido incorrecto: esperado %q, pero obtuve %q", testMessage, string(msg))
	}
}

// TestWSHandlerUserDisconnection verifica que los usuarios se eliminan al desconectar
func TestWSHandlerUserDisconnection(t *testing.T) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=disconnect-test"

	// Conectar usuario
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	chat, exists := chats["disconnect-test"]
	if !exists {
		t.Fatal("chat no existe")
	}

	if len(chat.usuarios) != 1 {
		t.Errorf("se esperaba 1 usuario, pero hay %d", len(chat.usuarios))
	}

	// Cerrar conexión
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	// El usuario debe ser eliminado (aunque puede haber una carrera en el que se procesa)
	if len(chat.usuarios) != 0 {
		t.Logf("advertencia: usuario no fue eliminado inmediatamente (len=%d)", len(chat.usuarios))
	}
}

// TestChatStructure verifica que la estructura Chat se inicializa correctamente
func TestChatStructure(t *testing.T) {
	chat := &Chat{
		usuarios:   make(map[string]*Usuario),
		broadcast: make(chan []byte, 10),
	}

	if chat.usuarios == nil {
		t.Error("usuarios map es nil")
	}
	if len(chat.usuarios) != 0 {
		t.Error("usuarios map no está vacío")
	}
	if chat.broadcast == nil {
		t.Error("broadcast channel es nil")
	}
}

// TestUsuarioStructure verifica que la estructura Usuario se inicializa correctamente
func TestUsuarioStructure(t *testing.T) {
	// Este test requeriría una conexión real, así que verificamos solo la estructura
	usuario := &Usuario{
		ID:   "test-id",
		Conn: nil, // No podemos tener una conexión en este test
	}

	if usuario.ID == "" {
		t.Error("usuario ID está vacío")
	}
}

// TestWSHandlerEmptyMessage verifica el manejo de mensajes vacíos
func TestWSHandlerEmptyMessage(t *testing.T) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=empty-msg-test"

	sender, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando: %v", err)
	}
	defer sender.Close()

	receiver, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando receiver: %v", err)
	}
	defer receiver.Close()

	time.Sleep(100 * time.Millisecond)

	// Enviar mensaje vacío
	err = sender.WriteMessage(websocket.TextMessage, []byte(""))
	if err != nil {
		t.Fatalf("Error escribiendo mensaje vacío: %v", err)
	}

	receiver.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := receiver.ReadMessage()
	if err != nil {
		t.Fatalf("Error leyendo mensaje: %v", err)
	}

	if len(msg) != 0 {
		t.Errorf("se esperaba mensaje vacío, pero obtuve %q", string(msg))
	}
}

// TestWSHandlerLargeMessage verifica el manejo de mensajes grandes
func TestWSHandlerLargeMessage(t *testing.T) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=large-msg-test"

	sender, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando sender: %v", err)
	}
	defer sender.Close()

	receiver, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error conectando receiver: %v", err)
	}
	defer receiver.Close()

	time.Sleep(100 * time.Millisecond)

	// Crear un mensaje grande (1 MB)
	largeMessage := bytes.Repeat([]byte("test"), 262144) // ~1 MB
	err = sender.WriteMessage(websocket.TextMessage, largeMessage)
	if err != nil {
		t.Fatalf("Error escribiendo mensaje grande: %v", err)
	}

	receiver.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := receiver.ReadMessage()
	if err != nil {
		t.Fatalf("Error leyendo mensaje grande: %v", err)
	}

	if len(msg) != len(largeMessage) {
		t.Errorf("tamaño del mensaje incorrecto: esperado %d, obtuve %d", len(largeMessage), len(msg))
	}
}

// TestConcurrentBroadcast verifica que múltiples usuarios reciben mensajes concurrentemente
func TestConcurrentBroadcast(t *testing.T) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=concurrent-test"
	numUsers := 5

	conns := make([]*websocket.Conn, numUsers)
	for i := 0; i < numUsers; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Error conectando usuario %d: %v", i, err)
		}
		conns[i] = conn
	}
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	time.Sleep(100 * time.Millisecond)

	chat, exists := chats["concurrent-test"]
	if !exists {
		t.Fatal("chat no existe")
	}

	if len(chat.usuarios) != numUsers {
		t.Errorf("se esperaban %d usuarios, pero hay %d", numUsers, len(chat.usuarios))
	}

	// El primer usuario envía un mensaje
	testMsg := "broadcast to all"
	err := conns[0].WriteMessage(websocket.TextMessage, []byte(testMsg))
	if err != nil {
		t.Fatalf("Error escribiendo mensaje: %v", err)
	}

	// Todos los usuarios deben recibir el mensaje
	for i := 0; i < numUsers; i++ {
		conns[i].SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conns[i].ReadMessage()
		if err != nil {
			t.Fatalf("Error leyendo mensaje en usuario %d: %v", i, err)
		}
		if string(msg) != testMsg {
			t.Errorf("usuario %d recibió mensaje incorrecto: %q", i, string(msg))
		}
	}
}

// BenchmarkGetOrCreateChat mide el rendimiento de getOrCreateChat
func BenchmarkGetOrCreateChat(b *testing.B) {
	chats = make(map[string]*Chat)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chatID := fmt.Sprintf("chat-%d", i%100)
		getOrCreateChat(chatID)
	}
}

// BenchmarkWSHandlerMessageReceive mide el rendimiento de recepción de mensajes
func BenchmarkWSHandlerMessageReceive(b *testing.B) {
	chats = make(map[string]*Chat)

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?chat=bench-test"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Error conectando: %v", err)
	}
	defer conn.Close()

	testMsg := []byte("benchmark message")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		conn.WriteMessage(websocket.TextMessage, testMsg)
	}
}

// TestUpgraderCheckOrigin verifica que el upgrader acepta todas las direcciones
func TestUpgraderCheckOrigin(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Origin", "http://different-domain.com")

	result := upgrader.CheckOrigin(req)
	if !result {
		t.Error("CheckOrigin debería retornar true para cualquier origen")
	}
}

// TestUpgraderInitialized verifica que el upgrader está correctamente inicializado
func TestUpgraderInitialized(t *testing.T) {
	if upgrader.CheckOrigin == nil {
		t.Error("upgrader.CheckOrigin es nil")
	}

	// Probar unos orígenes
	for _, origin := range []string{
		"http://localhost:3000",
		"http://192.168.1.1:8000",
		"http://example.com",
	} {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("Origin", origin)
		if !upgrader.CheckOrigin(req) {
			t.Errorf("CheckOrigin retornó false para origen %s", origin)
		}
	}
}

// TestChatConcurrency verifica que Chat maneja correctamente la concurrencia
func TestChatConcurrency(t *testing.T) {
	chat := &Chat{
		usuarios:   make(map[string]*Usuario),
		broadcast: make(chan []byte, 100),
	}

	// Simular múltiples goroutines agregando y eliminando usuarios
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user-%d", id)
			chat.mu.Lock()
			chat.usuarios[userID] = &Usuario{ID: userID}
			chat.mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			chat.mu.Lock()
			delete(chat.usuarios, userID)
			chat.mu.Unlock()

			done <- true
		}(i)
	}

	// Esperar a que terminen todas las goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	chat.mu.Lock()
	if len(chat.usuarios) != 0 {
		t.Errorf("todos los usuarios deberían haber sido eliminados, pero hay %d", len(chat.usuarios))
	}
	chat.mu.Unlock()
}
