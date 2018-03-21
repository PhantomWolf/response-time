package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		for i := 0; i <= 1000000; i++ {
			w.Write([]byte("aiosdjfoiauwe98r7298374uifuasjkfhas9udfy982734981274iojhfaskjdfh97239847"))
		}
	})
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
