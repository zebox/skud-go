package main

import (
	"bufio"

	"encoding/json"
	"io/ioutil"

	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/boltdb/bolt"
	"github.com/tarm/serial"
)

const dbname = "access.db"

var isOpen, isHLock bool = false, false
var serialPort *serial.Port

func main() {
	config, err := readConfig()

	if err != nil {
		fmt.Printf("Error read config file %s", err.Error())
		return
	}

	f, err := os.OpenFile(config.LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	http.HandleFunc("/"+config.NormalModeEndpoint, webNormalMode)
	http.HandleFunc("/"+config.HardLockModeEndpoint, webHLockMode)
	http.HandleFunc("/"+config.CloseEndpoint, webCloseRelay)
	http.HandleFunc("/"+config.OpenEndpoint, webOpenRelay)
	http.HandleFunc("/"+config.AddKeyEndpoint, addKey)
	http.HandleFunc("/"+config.ReadKeysEndpoint, readKeys)
	http.HandleFunc("/"+config.DeleteKeyEndpoint, deleteKey)

	go http.ListenAndServe(":"+config.HTTPPort, nil)

	log.Printf("Listening on port %s...", config.HTTPPort)

	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Close()

	// c := &serial.Config{Name: config.SerialPort, Baud: 9600}
	// s, err := serial.OpenPort(c)

	// if err != nil {
	// 	fmt.Printf("Error open serial port %s ", err.Error())
	// 	log.Fatal(err)

	// }
	// serialPort = s
	ch := make(chan bool) // wait chanel until key is valid
	//go getData(ch, s)

	for {
		time.Sleep(time.Second)
		tmp := <-ch
		if tmp {
			if isOpen {
				closeRelay()
			} else {
				openRelay()
			}
		}

	}

}

func getData(ch chan bool, s *serial.Port) {

	for {
		reader := bufio.NewReader(s)
		reply, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}
		k := string(reply)

		if chk := checkKey(k); chk {

			ch <- chk
			time.Sleep(2 * time.Second)
		}

	}

}
func invertBool() {
	isOpen = !isOpen
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func boltStore(value Key) {
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("keys"))
		if err != nil {
			return err
		}
		return b.Put([]byte(value.Key), []byte(value.isEnable))
	})
}

func boltRead(key string) bool {
	var strKey string
	db, err := bolt.Open(dbname, 0600, nil)

	if err != nil {
		log.Fatal(err)
		return false
	}

	defer db.Close()

	db.View(func(tx *bolt.Tx) error {

		re := regexp.MustCompile(`\r\n`)
		key := re.ReplaceAllString(key, "")
		re = regexp.MustCompile(`\n`)
		key = re.ReplaceAllString(key, "")
		re = regexp.MustCompile(`\r`)
		key = re.ReplaceAllString(key, "")
		log.Printf("Readed key: %s\n", key)

		b := tx.Bucket([]byte("keys"))
		v := b.Get([]byte(key))

		strKey = string(v)

		return nil
	})
	if strKey == "1" {
		log.Printf("Key %s is valid\n", key)
		return true
	}
	return false

}

func addKey(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	var key Key
	key.Key = params.Get("key")
	key.isEnable = params.Get("enable")
	boltStore(key)
	log.Printf("You add the key %s", key.Key)
	fmt.Fprintln(w, "You add the key", key.Key)

}
func readKeys(w http.ResponseWriter, r *http.Request) {
	keys := make(map[string]string)
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("keys"))

		b.ForEach(func(k, v []byte) error {
			keys[string(k)] = string(v)
			fmt.Printf("map: %s\n", keys[string(k)])
			return nil
		})
		return nil
	})
	data, _ := json.Marshal(keys)
	fmt.Fprintln(w, string(data))
}

func deleteKey(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	deleteKey := params.Get("key")
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("keys"))
		err := b.Delete([]byte(deleteKey))
		if err != nil {
			fmt.Printf("Key: \"%s\" delete failed: %s\n", deleteKey, err.Error())
			return err
		}
		fmt.Fprintf(w, "Key: \"%s\" deleted succesfully\n", deleteKey)

		// Persist bytes to users bucket.
		return nil
	})

}

func webNormalMode(w http.ResponseWriter, r *http.Request) {
	isHLock = false
	_, err := serialPort.Write([]byte("hlock0"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, "Normal Mode")
}
func webHLockMode(w http.ResponseWriter, r *http.Request) {
	_, err := serialPort.Write([]byte("hlock1"))
	if err != nil {
		log.Fatal(err)
	}
	isHLock = true
	fmt.Fprintln(w, "HardLock Mode")
}
func webCloseRelay(w http.ResponseWriter, r *http.Request) {
	switchRelay()
	fmt.Fprintln(w, "switch relay")
}
func webOpenRelay(w http.ResponseWriter, r *http.Request) {
	openRelay()
	fmt.Fprintln(w, "open lock")
}

func closeRelay() {

	_, err := serialPort.Write([]byte("close"))
	if err != nil {
		log.Fatal(err)
	}
	invertBool()
	log.Println("Close")

}

func openRelay() {

	_, err := serialPort.Write([]byte("open"))
	if err != nil {
		log.Fatal(err)
	}
	invertBool()
	log.Println("Open")

}
func switchRelay() {
	if isOpen {
		closeRelay()
	} else {
		openRelay()
	}
}
func checkKey(key string) bool {
	if boltRead(key) {

		return true
	}
	return false
}

func readConfig() (*Config, error) {
	plan, _ := ioutil.ReadFile("config.json")
	config := Config{}
	err := json.Unmarshal([]byte(plan), &config)
	return &config, err
}
