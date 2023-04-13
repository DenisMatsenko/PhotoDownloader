package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

type Cat struct {
	ID string `json:"id"`
	URL string `json:"url"`
	Width int `json:"width"`
	Height int `json:"height"`
}

type Request struct {
	Width  int
	Height int
	PhotosCount int
	SaveToFolder string
}

func main() {
	request := Request{}
	request.Width, request.Height = photoSizeInput()
	request.PhotosCount = photosCountInput()
	request.SaveToFolder = photosFolderInput()

	exec(request)
}

func exec(request Request) {
	channel := make(chan string, request.PhotosCount * 2)
	errorChannel := make(chan error, request.PhotosCount * 2)

	var wg sync.WaitGroup
	for i := 0; i < request.PhotosCount; i++ {
		wg.Add(1)
		go func() {
			downloadCatPhoto(request, channel, errorChannel)
			wg.Done()
		}()	
	}
	wg.Wait()

	close(channel)
	close(errorChannel)

	for err := range errorChannel {
		fmt.Println("Error: ", err)
	}

	for str := range channel {
		fmt.Println(str)
	}
}


func downloadCatPhoto(request Request, channel chan string, errorChannel chan error) {
	cat, err := getCat()
	if err != nil {
		errorChannel <- err
		return
	}

	os.Mkdir(request.SaveToFolder, 0777)
	var pathToSave string =  fmt.Sprintf("%s/cat-%s.jpg", request.SaveToFolder, cat.ID)

	file, err := os.Create(pathToSave)
	if err != nil {
		errorChannel <- err
		return
	}
	defer file.Close()


	img, err := getCatImage(*cat)
	if err != nil {
		errorChannel <- err
		return
	}

	_, err = io.Copy(file, img)
	if err != nil {
		errorChannel <- err
		return
	}

	channel <- fmt.Sprintf("Cat photo saved to %s", pathToSave)
}

func getCatImage(cat Cat) (io.ReadCloser, error) {
	resp, err := http.Get(cat.URL)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func getCat() (*Cat, error) {
	var catArray []Cat

	resp, err := http.Get("https://api.thecatapi.com/v1/images/search")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&catArray)
	if err != nil {
		return nil, err
	}
	return &catArray[0], nil
}

func photosFolderInput() string {
	var folderName string
	for {
		fmt.Print("Path: ")
		n, err := fmt.Scanf("%s", &folderName)
		if err != nil || n != 1 {
			fmt.Println("Invalid input")
			continue
		}
		break
	}
	return folderName
}

func photoSizeInput() (int, int) {
	var width, height int
	for {
		fmt.Print("Width: ")
		n, err := fmt.Scanf("%d", &width)
		if err != nil || width < 1 || n != 1 {
			fmt.Println("Invalid input")
			continue
		}
		break
	}
	for {
		fmt.Print("Height: ")
		n, err := fmt.Scanf("%d", &height)
		if err != nil || height < 1 || n != 1 {
			fmt.Println("Invalid input")
			continue
		}
		break
	}
	return width, height
}

func photosCountInput() int {
	var photosCount int
	for {
		fmt.Print("Photos count: ")
		n, err := fmt.Scanf("%d", &photosCount)
		if err != nil || photosCount < 1 || n != 1 {
			fmt.Println("Invalid input")
			continue
		}
		break
	}
	return photosCount
}

