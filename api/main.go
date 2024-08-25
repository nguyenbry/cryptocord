package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nguyenbry/crypto-reports/crypto"
)

func main() {
	key, ok := os.LookupEnv("CMC_KEY")
	if !ok {
		log.Fatal("no key")
	}

	c := crypto.New(key)

	res, err := c.Bitcoin(context.Background())

	if err != nil {
		log.Fatal("error occurred getting Bitcoin price", err)
	}

	fmt.Println(res)
}
