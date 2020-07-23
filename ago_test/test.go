package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
)

type server int

func (h *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	w.Write([]byte("Hello World"))
}

func main() {
	//var s server
	//http.ListenAndServe("localhost:9999", &s)
	data := []int{1,2,5,6,8,20,22,23,24,26,29,32}
	for k, v := range data {
		fmt.Println(k,v)
	}
	x := 23
	i := sort.Search(len(data), func(i int) bool { return data[i] <= x })
	fmt.Println(i)
	if i < len(data) && data[i] == x {
		// x is present at data[i]
	} else {
		// x is not present in data,
		// but i is the index where it would be inserted.
	}
}