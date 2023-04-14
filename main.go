package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/oliamb/cutter"
)

type Cat struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Image  image.Image
}

type Request struct {
	Width        int
	Height       int
	PhotosCount  int
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
	os.Mkdir(request.SaveToFolder, 0777)

	channel := make(chan string, request.PhotosCount*2)
	errorChannel := make(chan error, request.PhotosCount*2)

	var wg sync.WaitGroup
	for i := 0; i < request.PhotosCount; i++ {
		wg.Add(1)
		go func() {
			processCatPhotoRequest(request, channel, errorChannel)
			wg.Done()
		}()
	}
	wg.Wait()

	close(channel)
	close(errorChannel)

	isError := false
	for err := range errorChannel {
		isError = true
		fmt.Println("Error: ", err)
	}
	if isError {
		os.Exit(1)
	}

	for str := range channel {
		fmt.Println(str)
	}
}

func processCatPhotoRequest(request Request, channel chan string, errorChannel chan error) {
	cat, err := newCat()
	if err != nil {
		errorChannel <- err
		return
	}
	if cat == nil {
		channel <- "No cat photo found"
		return
	}

	img, err := cropImage(cat.Image, request)
	if err != nil {
		errorChannel <- err
		return
	}

	var pathToSave string = fmt.Sprintf("%s/cat-%s.jpg", request.SaveToFolder, cat.ID)

	imageFile, err := os.Create(pathToSave)
	if err != nil {
		errorChannel <- err
		return
	}
	defer imageFile.Close()

	err = jpeg.Encode(imageFile, img, nil)
	if err != nil {
		errorChannel <- err
		return
	}

	channel <- fmt.Sprintf("Cat photo saved to %s", pathToSave)
}

func newCat() (*Cat, error) {
	var catArray []Cat
	var cat *Cat

	attempts := 5
	for i := 0; i < attempts; i++ {
		resp, err := http.Get("https://api.thecatapi.com/v1/images/search")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP status code: %d", resp.StatusCode)
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&catArray)
		if err != nil {
			return nil, err
		}
		if len(catArray) != 1 {
			return nil, fmt.Errorf("expected 1 cat, got %d", len(catArray))
		}

		cat = &catArray[0]

		if strings.Contains(strings.ToLower(cat.URL), ".jpg") || 
			strings.Contains(strings.ToLower(cat.URL), ".jpeg") {
			break
		}

		fmt.Println("Not a jpg image, trying again...")

		if i == attempts-1 {
			fmt.Println("Failed to fetch cat photo")
			return nil, nil
		}
	}

	catUrlResp, err := http.Get(cat.URL)
	if err != nil {
		return nil, err
	}

	cat.Image, err = jpeg.Decode(catUrlResp.Body)
	if err != nil {
		return nil, err
	}

	return cat, nil
}

func cropImage(image image.Image, request Request) (image.Image, error) {
	croppedImgage, err := cutter.Crop(image, cutter.Config{
		Width:  request.Width,
		Height: request.Height,
		Mode:   cutter.Centered,
	})
	if err != nil {
		return nil, err
	}
	return croppedImgage, nil
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
