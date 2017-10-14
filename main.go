package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

const (
	targetURLEnv = "TARGET_URL"
)

var (
	mc        *memcache.Client
	targetURL string
)

func fetch(path string) *http.Response {

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}
	url := targetURL + path
	fmt.Printf("[%s] Fetching\n", url)

	response, err := netClient.Get(url)
	if err != nil {
		fmt.Printf("[%s] Error Fetching: %v", url, err)
		response = &http.Response{
			Status: "503",
		}
	} else {
		fmt.Printf("[%s] Cached\n", url)
	}

	return response
}

func req(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("[%s] Incoming Request\n", r.URL.Path)

	val, err := mc.Get(r.URL.Path)

	if err != nil {
		fmt.Printf("[%s] Cache MISS\n", r.URL.Path)
		fmt.Println(err)

		backendResponse := fetch(r.URL.Path)

		body, _ := ioutil.ReadAll(backendResponse.Body)

		w.Write(body)

		mc.Set(&memcache.Item{Key: r.URL.Path, Value: []byte(body)})
	} else {
		fmt.Printf("[%s] Cache HIT\n", r.URL.Path)

		w.Write(val.Value)
	}

}

func main() {

	targetURL = os.Getenv(targetURLEnv)
	if targetURL == "" {
		fmt.Printf("Set %s to the domain that will be cached", targetURLEnv)
	}

	mc = memcache.New("127.0.0.1:11211")

	http.HandleFunc("/", req)
	http.ListenAndServe(":8080", nil)
}
