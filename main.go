package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// CompilerResponse is a JSON message returned from the /vm/compile endpoint
type CompilerResponse struct {
	BSS, Raw string
}

var apiKey = getAPIKeyFromEnv()

func getAPIKeyFromEnv() string {
	key := os.Getenv("SF_API_KEY")
	if len(key) == 0 {
		log.Fatal("could not retrieve SF_API_KEY value from environment")
	}
	return key
}

func loadFile(location string) *os.File {
	file, err := os.Open(location)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func writeJSONTemplateToFile(fileNamePrefix, json, bss, data string) {
	json = strings.Replace(json, bss, "%s", 1)
	json = strings.Replace(json, data, "%s", 1)
	path := fmt.Sprintf("./%s.json", fileNamePrefix)
	fmt.Println("writing to " + path)
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	file.Write([]byte(json))
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

func compile() {
	if len(os.Args) != 4 {
		//i'll do proper flags later, promise
		log.Fatal("Need to specify an action (compile), C source file, and a filename prefix to write the .bss and .data files to")
	}

	source := loadFile(os.Args[2])
	fileNamePrefix := os.Args[3]

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
	writeJSONTemplateToFile(fileNamePrefix, string(body), cr.BSS, cr.Raw)
}

func generateTarget(jsonTemplate string, bssFile, dataFile *os.File) io.Reader {
	bssFileInfo, err := bssFile.Stat()
	if err != nil {
		log.Fatal(err)
	}
	dataFileInfo, err := dataFile.Stat()
	if err != nil {
		log.Fatal(err)
	}
	bssBytes := make([]byte, bssFileInfo.Size())
	dataBytes := make([]byte, dataFileInfo.Size())
	_, err = bssFile.Read(bssBytes)
	if err != nil {
		log.Fatal(err)
	}
	_, err = dataFile.Read(dataBytes)
	if err != nil {
		log.Fatal(err)
	}
	bssBase64 := base64.StdEncoding.EncodeToString(bssBytes)
	dataBase64 := base64.StdEncoding.EncodeToString(dataBytes)

	escapedTemplate := fmt.Sprintf(jsonTemplate, bssBase64, dataBase64)
	return strings.NewReader(escapedTemplate)
}

func write() {
	if len(os.Args) != 5 {
		//i'll do proper flags later, promise
		log.Fatal("Need to specify an action (write), the .json template, the .bss and .data source files")
	}

	///todo: use ioutil.ReadFile for all file reads
	jsonBytes, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	bssFile := loadFile(os.Args[3])
	defer bssFile.Close()
	dataFile := loadFile(os.Args[4])
	defer dataFile.Close()

	target := generateTarget(string(jsonBytes), bssFile, dataFile)

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.stockfighter.io/trainer/vm/write", target)
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
	//todo: yeah this is lazy. Should derserialize the JSON
	fmt.Println(string(body))
}

func main() {
	action := os.Args[1]
	if action == "compile" {
		compile()
	} else if action == "write" {
		write()
	} else {
		log.Fatal("available actions: compile or write")
	}
}
