package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const (
	defaultHTTPPort  = "8080"
	httpPortEnv      = "HTTP_PORT"
	defaultHTTPSPort = "4433"
	httpsPortEnv     = "HTTPS_PORT"
	certPathEnv      = "CERT_PATH"
	certKeyPathEnv   = "CERT_KEY_PATH"
	timeFormat       = "01-02-2006 03:04:05PM"
)

var (
	httpPort    string
	httpsPort   string
	certPath    string
	certKeyPath string
	db          *bolt.DB
)

func updateKey(bucketName []byte, key []byte, value []byte) {
	err := db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		err = bkt.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("[%s] Incoming GET Request\n", r.URL.Path)

	var val []byte
	s := strings.Split(r.URL.Path, "/")
	bucketName := []byte(s[1])
	key := []byte(s[2] + "/" + s[3])
	err := db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketName)
		if bkt != nil {
			val = bkt.Get(key)
		}
		return nil
	})

	if err != nil || len(val) == 0 {
		fmt.Printf("[%s] Cache MISS\n", r.URL.Path)
		fmt.Println(err)

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Status: 303"))
	} else {
		fmt.Printf("[%s] Cache HIT\n", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		w.Write(val)
	}

}

func handlePut(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("[%s] Incoming PUT Request\n", r.URL.Path)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	s := strings.Split(r.URL.Path, "/")
	bucketName := []byte(s[1])
	key := []byte(s[2] + "/" + s[3])

	if strings.Contains(r.URL.Path, "/clear") {
		clearKey := strings.Split(r.URL.Path, s[1])[1]
		clearKey = strings.Split(clearKey, "/")[1]
		clearAllPath(bucketName, []byte(clearKey+"/all"))
	}
	updateKey(bucketName, key, []byte(body))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Status: 200"))
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	reqMethod := strings.ToUpper(r.Method)
	fmt.Printf("[%s] Incoming %s Request\n", r.URL.Path, reqMethod)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, reqErr := http.NewRequest(reqMethod, r.URL.Path, r.Body)
	if reqErr != nil {
		fmt.Printf("reqErr: %v\n", reqErr)
		w.Write([]byte("Status: 503"))
		return
	}

	for k, v := range r.Header {
		req.Header.Add(k, strings.Join(v, ", "))
	}

	resp, resErr := client.Do(req)
	if resErr != nil {
		fmt.Printf("resErr: %v\n", resErr)
		w.Write([]byte("Status: 503"))
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
	fmt.Printf("[%s] %s Response: %s\n", r.URL.Path, reqMethod, string(body))
}

func clearAllPath(bucketName []byte, key []byte) {
	err := db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketName)
		if bkt != nil {
			err := bkt.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defVal
}

func main() {

	httpPort = getEnv(httpPortEnv, defaultHTTPPort)
	httpsPort = getEnv(httpsPortEnv, defaultHTTPSPort)
	certPath = getEnv(certPathEnv, "")
	certKeyPath = getEnv(certKeyPathEnv, "")
	switch {
	case certPath == "":
		fmt.Printf("Set %s to the path of the SSL .crt file\n", certPathEnv)
		return
	case certKeyPath == "":
		fmt.Printf("Set %s to the path of the SSL .key file\n", certKeyPathEnv)
		return
	}

	db, _ = bolt.Open("db", 0600, nil)
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGet(w, r)
		} else if r.Method == http.MethodPut {
			handlePut(w, r)
		} else {
			handleReq(w, r)
		}
	})

	go func() {
		fmt.Printf("Listening on port: %v\n", httpPort)
		httpErr := http.ListenAndServe(":"+httpPort, nil)
		if httpErr != nil {
			fmt.Printf("Error starting HTTP server: %v\n", httpErr)
		}
	}()

	config := &tls.Config{}
	cert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
	if err != nil {
		fmt.Printf("Error loading certs: %v\n", err)
	}
	config.Certificates = append(config.Certificates, cert)
	config.BuildNameToCertificate()

	server := http.Server{
		Addr:      ":" + httpsPort,
		TLSConfig: config,
	}

	fmt.Printf("Listening on port: %v\n", httpsPort)
	httpsErr := server.ListenAndServeTLS("", "")
	if httpsErr != nil {
		fmt.Printf("Error starting HTTPS server: %v\n", httpsErr)
	}
}
