package main

import (
	"encoding/json"
	"fmt"
	"github.com/oliamb/cutter"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"sync"
)

type Cat struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Request struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	PhotosCount  int    `json:"photos_count"`
	SaveToFolder string `json:"save_to_folder"`
}

func main() {
	request := Request{}
	request.Width, request.Height = photoSizeInput()
	request.PhotosCount = photosCountInput()
	request.SaveToFolder = photosFolderInput()

	exec(request)
}

func exec(request Request) {
	channel := make(chan string, request.PhotosCount*2)
	errorChannel := make(chan error, request.PhotosCount*2)

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

	img, err := cat.getCropedImage(request, channel, errorChannel)
	if err != nil {
		errorChannel <- err
		return
	}

	os.Mkdir(request.SaveToFolder, 0777)
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

func getCat() (*Cat, error) {
	var catArray []Cat

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
	return &catArray[0], nil
}

func (cat *Cat) getCropedImage(request Request, channel chan string, errorChannel chan error) (image.Image, error) {
	resp, err := http.Get(cat.URL)
	if err != nil {
		return nil, err
	}

	img, err := cropImage(resp.Body, request, channel, errorChannel)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func cropImage(resp io.ReadCloser, request Request, channel chan string, errorChannel chan error) (image.Image, error) {
	originalImage, err := jpeg.Decode(resp)
	if err != nil {
		downloadCatPhoto(request, channel, errorChannel)
		return nil, err
	}

	croppedImg, err := cutter.Crop(originalImage, cutter.Config{
		Width:  request.Width,
		Height: request.Height,
		Mode:   cutter.Centered,
	})
	if err != nil {
		return nil, err
	}
	return croppedImg, nil
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
