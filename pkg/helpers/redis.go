package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)



type DocAlreadyExists struct {
	Key		string
	Value	string
}

func (d *DocAlreadyExists) Error() string {
	return fmt.Sprintf("Key: '%s' already exists with value: '%s'", d.Key, d.Value)
}

type DocDoesntExist struct {
	Key		string
}

func (d *DocDoesntExist) Error() string {
	return fmt.Sprintf("Document with ID: '%s' does not exist.", d.Key)
}

type InvalidTopic struct {Topic string}

func (i *InvalidTopic) Error() string {
	return fmt.Sprintf("Topic: %s is not a valid topic category.", i.Topic)
	}

type RedisConf struct {
	Addr	string
	Port	string
}

type RedisCaller struct {
	ctx 	context.Context
	Client	*redis.Client
}


/*
Creates a new RedisCaller struct
	:param redisCfg: a redis configuration struct
*/
func NewRedisClient(redisCfg RedisConf) *RedisCaller {
    return &RedisCaller{Client: redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%v", redisCfg.Addr, redisCfg.Port),
        DB:       0,  // use default DB
    }),
	ctx: context.Background(),}

}

/*
retrieves all of the document IDs in the Redis database
*/
func (r *RedisCaller) AllDocIds() ([]string, error) {
	return r.Client.Keys(r.ctx, "*").Result()
}

/*
Sets the item (id) to the value supplied in value
	:param doc: the documents.Document struct to input to the database
*/
func (r *RedisCaller) AddDoc(doc Document) error {
	_, ok := TopicMap[doc.Category]
	if !ok {
		return &InvalidTopic{Topic: doc.Category}
	}

	val, err := r.Client.Get(r.ctx, doc.Ident).Result()
	if err == redis.Nil {
		data, err := json.Marshal(&doc)
		if err != nil {
			return err
		}

		err = r.Client.Set(r.ctx, doc.Ident, data, 0).Err()
		if err != nil {
			return err
		}
		return nil
    } else if err != nil {
        return err
    }
	return &DocAlreadyExists{Key: doc.Ident, Value: val}


}


/*
Gets the item stored at the key (id)
	:param id: the id of the object to get
*/
func (r *RedisCaller) GetItem(id string) (*Document, error) {

	var doc Document
	val, err := r.Client.Get(r.ctx, id).Result()
	if err == redis.Nil {
		return nil, err
    } else if err != nil {
        return nil, err
    }
	data := []byte(val)
	err = json.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

/*
Delete the target document in redis
	:param id: the id to delete from redis
*/
func (r *RedisCaller) DeleteDoc(id string) error {
	_, err := r.Client.Get(r.ctx, id).Result()
	if err == redis.Nil {
		return &DocDoesntExist{id}
    } else if err != nil {
        return err
    }

	err = r.Client.Del(r.ctx, id).Err()
	if err != nil {
		return err
	}
	return nil
}

/*
Update a value in redis
	:param id: the id of the document to edit
*/
func (r *RedisCaller) editVal(id string, in interface{}) error {
	_, err := r.Client.Get(r.ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return &DocDoesntExist{Key: id}
		}
		return err
	}

		data, err := json.Marshal(&in)
		if err != nil {
			return err
		}

		err = r.Client.Set(r.ctx, id, data, 0).Err()
		if err != nil {
			return err
		}
		return nil
    }

func (r *RedisCaller) SeedData(seedLoc string) error {
	dirs, err := os.ReadDir(seedLoc)
	if err != nil {
		return err
	}
	for i := range dirs {
		key := strings.Split(dirs[i].Name(), ".")[0]
		b, err := os.ReadFile(fmt.Sprintf("%s/%s", seedLoc, dirs[i].Name()))
		if err != nil {
			return err
		}
		err = r.Client.Set(r.ctx, key, b, 0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}





func (r *RedisCaller) UpdatePost(id string, new Document) error {
	return r.editVal(id, new)
}

func (r *RedisCaller) UpdateHeader(id string, new HeaderCollection) error {
	return r.editVal(id, new)
}

func (r *RedisCaller) UpdateMenu(id string, new MenuElement) error {
	return r.editVal(id, new)
}