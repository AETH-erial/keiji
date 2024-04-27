package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type InvalidSkipArg struct {Skip int}

func (i *InvalidSkipArg) Error() string {
	return fmt.Sprintf("Invalid skip amount was passed: %v", i.Skip)
}


type ImageStoreItem struct {
	Identifier		string	`json:"identifier"`
	Filename		string	`json:"filename"`
	AbsolutePath	string	`json:"absolute_path"`
	Title			string	`json:"title" form:"title"`
	Created			string	`json:"created"`
	Desc			string	`json:"description" form:"description"`
	Category		string	`json:"category"`
	ApiPath			string
}

/*
Create a new ImageStoreItem
	:param fname: the name of the file to be saved
	:param title: the canonical title to give the image
	:param desc: the description to associate to the image
*/
func NewImageStoreItem(fname string, title string, desc string) *ImageStoreItem {
	id := uuid.New()
	img := ImageStoreItem{
		Identifier: id.String(),
		Filename: fname,
		Title: title,
		Category: DIGITAL_ART,
		AbsolutePath: fmt.Sprintf("%s/%s", GetImageStore(), fname),
		Created: time.Now().UTC().String(),
		Desc: desc,
	}
	return &img
}


/*
Function to return the location of the image store. Wrapping the env call in
a function so that refactoring is easier
*/
func GetImageStore() string {
	return os.Getenv(env.IMAGE_STORE)
}

/*
Return database entries of the images that exist in the imagestore
	:param rds: pointer to a RedisCaller to perform the lookups with
*/
func GetImageData(rds *RedisCaller) ([]*ImageStoreItem, error) {
	ids, err := rds.GetByCategory(DIGITAL_ART)
	if err != nil {
		return nil, err
	}

	var imageEntries []*ImageStoreItem
	for i := range ids {
		val, err := rds.Client.Get(rds.ctx, ids[i]).Result()
		if err == redis.Nil {
			return nil, err
		} else if err != nil {
			return nil, err
		}
		data := []byte(val)
		var imageEntry ImageStoreItem
		err = json.Unmarshal(data, &imageEntry)
		if err != nil {
			return nil, err
		}
		imageEntry.ApiPath = fmt.Sprintf("/api/v1/images/%s", imageEntry.Filename)
		imageEntries = append(imageEntries, &imageEntry)
	}
	return imageEntries, err
}

