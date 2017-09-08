package webapp

/*
import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

var ticker *time.Ticker
var dirsList []string = []string{"up", "left", "down", "right"}

type WebClientData struct {
	websockPort string
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	//TODO create login page
	fmt.Fprintf(w, "Hi there, welcome to BlobTags!")
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	h := WebClientData{"7000"}
	t.Execute(w, h)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[1:]
	fmt.Println(" -- staticHandler; Searching for: " + filePath)
	http.ServeFile(w, r, filePath)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	//filePath := r.URL.Path[1:]
	fmt.Println(" -- Query[\"dir\"]: " + r.URL.Query().Get("dir"))
	fmt.Fprintf(w, "THX for the update!")
}

func main() {
	http.HandleFunc("/", baseHandler)
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/go", goHandler)
	http.HandleFunc("/update/", updateHandler)

	go func() {
		fmt.Println(" Setving http..")
		http.ListenAndServe(":8080", nil)
	}()

	http.Handle("/subscribe", websocket.Handler(onWebSocketConnect))
	fmt.Println(" Serving ws..")

	go func() {
		err := http.ListenAndServe(":7000", nil)
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	ticker = time.NewTicker(3 * time.Second)

	select {}
}

func onWebSocketConnect(ws *websocket.Conn) {

	msg := make([]byte, 512)
	_, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received from ws: %s\n", msg)

	_, err = ws.Write([]byte("Hello!"))
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ticker.C:
			randDir := dirsList[rand.Intn(4)]
			//TODO catch write error
			_, err = ws.Write([]byte(randDir))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sending to ws: %s\n", msg)
		}

	}
}

*/
