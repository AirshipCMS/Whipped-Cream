package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// TTL info for each request
type TTL struct {
	CachedTime time.Time
	ExpiryTime time.Time
}

const (
	defaultPort        = "8080"
	portEnv            = "PORT"
	targetURLEnv       = "TARGET_URL"
	defaultMemcacheURL = "127.0.0.1:11211"
	memcacheURLEnv     = "MEMCACHE_URL"
	defaultTTL         = "3600" // seconds
	ttlEnv             = "TTL"
	timeFormat         = "01-02-2006 03:04:05PM"
)

var (
	mc          *memcache.Client
	reqMap      = make(map[string]TTL)
	port        string
	targetURL   string
	memcacheURL string
	ttl         string
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

func cacheMiss(urlPath string) []byte {
	backendResponse := fetch(urlPath)

	body, _ := ioutil.ReadAll(backendResponse.Body)

	mc.Set(&memcache.Item{Key: urlPath, Value: []byte(body)})

	d, _ := time.ParseDuration(ttl + "s")
	reqMap[urlPath] = TTL{
		CachedTime: time.Now().Local(),
		ExpiryTime: time.Now().Local().Add(time.Duration(d)),
	}

	fmt.Printf("Expiry Time: %s\n", reqMap[urlPath].ExpiryTime.Format(timeFormat))
	return body
}

func req(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("[%s] Incoming Request\n", r.URL.Path)

	val, err := mc.Get(r.URL.Path)

	if err != nil {
		fmt.Printf("[%s] Cache MISS\n", r.URL.Path)
		fmt.Println(err)

		w.Write(cacheMiss(r.URL.Path))
	} else if time.Now().Local().After(reqMap[r.URL.Path].ExpiryTime) {
		w.Write(cacheMiss(r.URL.Path))
	} else {
		fmt.Printf("[%s] Cache HIT\n", r.URL.Path)

		w.Write(val.Value)
	}

}

func getEnv(key, defVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defVal
}

func main() {

	port = getEnv(portEnv, defaultPort)
	memcacheURL = getEnv(memcacheURLEnv, defaultMemcacheURL)
	ttl = getEnv(ttlEnv, defaultTTL)
	targetURL = os.Getenv(targetURLEnv)
	if targetURL == "" {
		fmt.Printf("Set %s to the domain that will be cached\n", targetURLEnv)
	}

	mc = memcache.New(memcacheURL)

	http.HandleFunc("/", req)
	fmt.Printf("Listening on port: %v\n", port)
	http.ListenAndServe(":"+port, nil)
}
