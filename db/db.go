package db

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
)

type DBKeys struct {
	Outpath string `json:"outpath"`
	Jspath  string `json:"jspath"`
}

var (
	bucketName = []byte("paths")
	dbName     = []byte("twitchdl.db")
)

func SetupDB() (*bolt.DB, error) {
	// p, err := utils.IncludeExecPath(string(dbName))
	// if err != nil {
	// 	return nil, err
	// }
	db, err := bolt.Open(string(dbName), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could no open db: %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("error while creating the bucket: %w", err)
		}
		return nil
	})
	return db, err
}

func PrintConfig(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		b.ForEach(func(k, v []byte) error {
			fmt.Println(string(k), string(v))
			return nil
		})
		return nil
	})
}

func UpdateBucketValues(db *bolt.DB, dbkeys DBKeys) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		values := reflect.ValueOf(dbkeys)
		keys := reflect.TypeOf(dbkeys)
		for i := 0; i < keys.NumField(); i++ {
			value := values.Field(i)
			key := keys.Field(i).Tag.Get("json")
			// UpdateKeyValue(db, key, value.String())
			b.Put([]byte(key), []byte(value.String()))
		}
		return nil
	})
}

func GetBucketValues(db *bolt.DB) (*DBKeys, error) {
	data := make(map[string]string)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		err := b.ForEach(func(k, v []byte) error {
			data[string(k)] = string(v)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	marshaledBytes, err := json.Marshal(data)
	var keys *DBKeys
	if err := json.Unmarshal(marshaledBytes, &keys); err != nil {
		return nil, err
	}
	return keys, err
}

func GetBucketValue(db *bolt.DB, key string) (string, error) {
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket(bucketName).Get([]byte(key))
		if len(value) == 0 {
			return fmt.Errorf("no value for key: %s", key)
		}
		return nil
	})
	return string(value), err
}

func InitPathValues(db *bolt.DB) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		b.Put([]byte("outpath"), []byte("./"))
		b.Put([]byte("jspath"), []byte("./scripts"))
		return nil
	})
}

// func UpdateKeyValue(db *bolt.DB, key, v string) error {
// 	fmt.Println("Updating", key, v)
// 	return db.Update(func(tx *bolt.Tx) error {
// 		b := tx.Bucket(bucketName)
// 		err := b.Put([]byte(key), []byte(v))
// 		return err
// 	})
// }
