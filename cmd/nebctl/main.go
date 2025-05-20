package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	wasmFile := flag.String("wasm", "", "path to wasm")
	name := flag.String("name", "", "function name")
	addr := flag.String("addr", "http://localhost:8080", "edge address base URL")
	flag.Parse()
	if *wasmFile == "" || *name == "" {
		log.Fatal("usage: nebctl -wasm double.wasm -name double")
	}
	data, err := ioutil.ReadFile(*wasmFile)
	if err != nil {
		log.Fatal(err)
	}
	payload := map[string]string{
		"name": *name,
		"wasm": base64.StdEncoding.EncodeToString(data),
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(*addr+"/deploy", "application/json", bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("deploy failed: %s", resp.Status)
	}
	fmt.Println("deployed", *name, "to", *addr)
}
