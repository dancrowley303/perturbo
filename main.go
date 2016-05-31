package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// CompilerResponse is a JSON message returned from the /vm/compile endpoint
type CompilerResponse struct {
	BSS, Raw string
}

func getAPIKeyFromEnv() string {
	key := os.Getenv("SF_API_KEY")
	if len(key) == 0 {
		log.Fatal("could not retrieve SF_API_KEY value from environment")
	}
	return key
}

func loadSourceFile(location string) *os.File {
	file, err := os.Open(location)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func writeBSSToFile(fileNamePrefix, bss string) {
	path := fmt.Sprintf("./%s.bss", fileNamePrefix)
	fmt.Println("writing to " + path)
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := base64.StdEncoding.DecodeString(bss)
	if err != nil {
		log.Fatal(err)
	}
	file.Write(bytes)
}

func writeDataToFile(fileNamePrefix, data string) {
	path := fmt.Sprintf("./%s.data", fileNamePrefix)
	fmt.Println("writing to " + path)
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}
	file.Write(bytes)
}

func main() {
	apiKey := getAPIKeyFromEnv()
	if len(os.Args) != 3 {
		log.Fatal("Need to specify a C source file, and a filename prefix to write the .bss and .data files to")
	}

	source := loadSourceFile(os.Args[1])
	fileNamePrefix := os.Args[2]

	defer source.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.stockfighter.io/trainer/vm/compile", source)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("X-Starfighter-Authorization", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var cr CompilerResponse

	if err := json.Unmarshal(body, &cr); err != nil {
		log.Fatal(err)
	}

	writeBSSToFile(fileNamePrefix, cr.BSS)
	writeDataToFile(fileNamePrefix, cr.Raw)

}
