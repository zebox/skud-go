package main

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const bucketName string = "keys"

func main() {
	// argsWithProg := os.Args
	// argsWithoutProg := os.Args[1:]

	fmt.Printf("cmd lenght %d\n", len(os.Args))
	if len(os.Args) > 1 {
		arg1 := os.Args[1]
		switch arg1 {
		case "read-keys":
			readKeys()
		case "delete-key":
			if len(os.Args) > 2 {
				arg2 := os.Args[2]
				deleteKey(arg2)
			} else {
				fmt.Println("Delete command has not a key value for execute")
			}
		case "add-key":
			if len(os.Args) > 2 {
				arg2 := os.Args[2]
				addKey(arg2)
			} else {
				fmt.Println("Add command has not a key value for execute")
			}
		default:
			printHelp()
		}
	} else {
		printHelp()
	}

	// if len(os.Args) > 2 {
	// 	arg2 := os.Args[2]
	// 	fmt.Println(arg2)
	// }
	// // fmt.Println(argsWithProg)
	// // fmt.Println(argsWithoutProg)
	// fmt.Println(arg1)
}

func readKeys() {
	db, err := bolt.Open("access.bolt", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucketName))

		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s, value=%s\n", k, v)
			return nil
		})
		return nil
	})
}

func deleteKey(key string) {
	db, err := bolt.Open("access.bolt", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte(bucketName))
		err := b.Delete([]byte(key))
		if err != nil {
			fmt.Printf("Key: \"%s\" delete failed: %s\n", key, err.Error())
			return err
		}
		fmt.Printf("Key: \"%s\" deleted succesfully\n", key)

		// Persist bytes to users bucket.
		return nil
	})
}

func addKey(key string) {
	db, err := bolt.Open("access.bolt", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(key), []byte("1"))
		if err != nil {
			fmt.Printf("Key: \"%s\" add failed: %s\n", key, err.Error())
			return err
		}
		fmt.Printf("Key: \"%s\" added succesfully\n", key)

		// Persist bytes to users bucket.
		return nil
	})
}

func printHelp() {
	fmt.Println("Help to keytools for SKUD Conroller\n")
	fmt.Println("read-keys  | Hasn't any arguments. Return All keys stored in Database\n")
	fmt.Println("delete-key  <key>| Has a key value argument for delete. Return error if failed\n")
	fmt.Println("add-key <key> | Has a key value argument for add. Return error if failed\n")
	fmt.Println()
	fmt.Println("Developer info@devgun.ru 2017\n")
}
