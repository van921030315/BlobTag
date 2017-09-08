package webapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/net/websocket"

	"time"

	"github.com/blobtag/dht"
	gorillaws "github.com/gorilla/websocket"
)

const WEBAPP_FOLDER = "../src/github.com/blobtag/webapp/"

// const WEBAPP_FOLDER = "Contents/Resources/"

type WebClientData struct {
	WebsockPort string
}

type FormData struct {
	RpcPort    int
	PlayerName string
}

type Header struct {
	Head1 string
	Head2 string
	Head3 string
}

func (data WebClientData) String() string {
	return "<WebsockPort: " + data.WebsockPort + ">"
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	//TODO create login page
	http.ServeFile(w, r, WEBAPP_FOLDER+"static/join.html")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := WEBAPP_FOLDER + r.URL.Path
	http.ServeFile(w, r, filePath)
}

type WebServer struct {
	WebServerPort  string
	WebsocketPort  string
	FromClientChan chan string
	ToClientChan   chan string
	SelfNodeId     int
	SelfBlobInfo   dht.BlobInfo
	FormDataChan   chan FormData
}

func (wa WebServer) WebsocketHandle(conn *gorillaws.Conn) {
	for {
		select {
		case msg := <-wa.ToClientChan:
			err := conn.WriteMessage(gorillaws.BinaryMessage, []byte(msg))
			if err != nil {
				fmt.Println("Websock write error: " + err.Error())
			}
		}
	}

}

func (wa WebServer) InitServer(wsPort string) {
	wa.WebServerPort = wsPort

	http.HandleFunc("/static/", staticHandler)

	http.HandleFunc("/websocketport", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", wa.WebsocketPort)
	})

	http.HandleFunc("/selfnodeinfo", func(w http.ResponseWriter, r *http.Request) {
		blobInfoJson, err := json.Marshal(wa.SelfBlobInfo)
		if err != nil {
			fmt.Println("!!! Error Marshalling self blob info: " + err.Error())
		}
		fmt.Fprintf(w, string(blobInfoJson))
	})

	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		//wait for 2 secs, then redirect to /go
		err := r.ParseForm()
		if err != nil {
			http.ServeFile(w, r, WEBAPP_FOLDER+"static/game.html")
		}
		playerName := r.Form["playername"][0]
		portRpc, err1 := strconv.Atoi(r.Form["portRpc"][0])

		if err1 != nil || !isValidPort(portRpc) || len(playerName) < 1 {
			fmt.Println("Inalid params")
			http.Redirect(w, r, "/", 301)
		}
		fmt.Printf("Got playername: %v\n", r.Form["playername"])
		//send to main thread
		wa.FormDataChan <- FormData{PlayerName: playerName, RpcPort: portRpc}
		wa.SelfBlobInfo.PlayerName = playerName
		time.Sleep(2 * time.Second)
		http.Redirect(w, r, "/go", 301)
	})

	http.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, WEBAPP_FOLDER+"static/game.html")
	})

	http.HandleFunc("/nodestateupdate/", func(w http.ResponseWriter, r *http.Request) {
		newState := r.URL.Query().Get("state")
		fmt.Fprintf(w, "THX for the update!")
		wa.FromClientChan <- newState
	})

	http.HandleFunc("/", baseHandler)

	go func() {
		fmt.Println(" Serving http..")
		http.ListenAndServe(":"+wsPort, nil)
	}()

	//initialize handler for websocket connection
	//Old websock connect

	http.Handle("/subscribe", websocket.Handler(func(ws *websocket.Conn) {
		for {
			select {
			case msg := <-wa.ToClientChan:
				_, err := ws.Write([]byte(msg))
				if err != nil {
					fmt.Println("Websock error: " + err.Error())
				}
			}
		}
	}))

	//gorilla version
	// http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
	// 	conn, err := gorillaws.Upgrade(w, r, w.Header(), 1024, 1024)
	// 	if err != nil {
	// 		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	// 	}
	// 	go wa.WebsocketHandle(conn)
	// })
	fmt.Println(" Serving ws..")

	go func() {
		err := http.ListenAndServe(":"+wa.WebsocketPort, nil)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()
}

func isValidPort(port int) bool {
	return 0 <= port && port < 65536
}
