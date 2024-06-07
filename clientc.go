package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

type ChunkLocation struct {
	SlaveURL string `json:"slave_url"`
	ChunkNum int    `json:"chunk_num"`
}

func main() {
	// Get the chunk count from the user
	var chunkCount int
	fmt.Print("Enter chunk count: ")
	fmt.Scan(&chunkCount)

	// Get the locations of the chunks
	locationsURL := fmt.Sprintf("http://192.168.1.5:33000/?chunks=%d", chunkCount)
	println(locationsURL)
	locationsResp, err := http.Get(locationsURL)
	if err != nil {
		fmt.Printf("Error getting chunk locations: %s\n", err)
		os.Exit(1)
	}
	defer locationsResp.Body.Close()
	
	// Parse the chunk locations from the response
	locationsData, err := ioutil.ReadAll(locationsResp.Body)
	// Decoding byte array to string
	str := string(locationsData)
	// Printing the decoded string
	fmt.Println(str)

	if err != nil {
		fmt.Printf("Error reading chunk locations: %s\n", err)
		os.Exit(1)
	}
	chunkLocations := make([]ChunkLocation, 0, chunkCount)
	err = json.Unmarshal(locationsData, &chunkLocations)
	if err != nil {
		fmt.Printf("Error decoding chunk locations: %s\n", err)
		os.Exit(1)
	}
	
	// Download the chunks from the slaves
	var wg sync.WaitGroup
	chunkData := make([][]byte, chunkCount)
	for _, chunkLocation := range chunkLocations {
		wg.Add(1)
		go func(chunkLocation ChunkLocation) {
			defer wg.Done()
	
			chunkURL := fmt.Sprintf("%s/?chunk=%d", chunkLocation.SlaveURL, chunkLocation.ChunkNum)
			chunkResp, err := http.Get(chunkURL)
			if err != nil {
				fmt.Printf("Error getting chunk %d from %s: %s\n", chunkLocation.ChunkNum, chunkLocation.SlaveURL, err)
				return
			}
			defer chunkResp.Body.Close()
	
			chunkData[chunkLocation.ChunkNum], err = ioutil.ReadAll(chunkResp.Body)
			if err != nil {
				fmt.Printf("Error reading chunk %d from %s: %s\n", chunkLocation.ChunkNum, chunkLocation.SlaveURL, err)
				return
			}
		}(chunkLocation)
	}
	wg.Wait()
	
	// Combine the chunks into a single file
	var combinedData []byte
	for _, chunk := range chunkData {
		combinedData = append(combinedData, chunk...)
	}
	
	// Write the combined data to a file
	err = ioutil.WriteFile("combined.bin", combinedData, 0644)
	if err != nil {
		fmt.Printf("Error writing combined file: %s\n", err)
		os.Exit(1)
	}
	
	fmt.Println("File downloaded and combined successfully.")
}