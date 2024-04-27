package controller

import (
	"fmt"
	"log"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
)

type Controller struct{
	WebRoot		string
	Domain		string
	RedisConfig helpers.RedisConf
	Cache		*helpers.AllCache
}

/*
Retrieve the header configuration from redis
*/
func (c *Controller) Headers() *helpers.HeaderCollection {
	headers, err := helpers.GetHeaders(c.RedisConfig)
	if err != nil {
		log.Fatal(err, "Headers couldnt be loaded. Exiting.")
	}
	return headers
}

/*
Retrieve the menu configuration from redis
*/
func (c *Controller) Menu() *helpers.MenuElement {
	links, err := helpers.GetMenuLinks(c.RedisConfig)
	if err != nil {
		log.Fatal(err, "Menu links couldnt be couldnt be loaded. Exiting.")
	}
	return links
}

/*
Retrieve the administrator table configuration from redis
*/
func (c *Controller) AdminTables() *helpers.AdminTables {
	tables, err := helpers.GetAdminTables(c.RedisConfig)
	if err != nil {
		log.Fatal(err, "Administrator tables couldnt be couldnt be loaded. Exiting.")
	}
	return tables
}


/*
Retrieve the post data and format it for the post management page
*/
func (c *Controller) FormatDocTable() *helpers.AdminTables {
	var postTables helpers.AdminTables
	for i := range helpers.Topics {
		docs, err := helpers.GetAllDocuments(helpers.Topics[i], c.RedisConfig)
		if err != nil {
			log.Fatal(err, "Post tables couldnt be couldnt be loaded. Exiting.")
		}
		var categoryTable helpers.Table
		categoryTable.TableName = helpers.Topics[i]
		for idx := range docs {
			categoryTable.TableData = append(categoryTable.TableData, helpers.TableData{
				DisplayName: docs[idx].Ident,
				Link: fmt.Sprintf("/admin/posts/%s", docs[idx].Ident),
			})
		}
		postTables.Tables = append(postTables.Tables, categoryTable)
	}
	return &postTables

}


/*
Save a new image store item
*/
func (c *Controller) SaveImage(img *helpers.ImageStoreItem) error {
	rds := helpers.NewRedisClient(c.RedisConfig)
	err := rds.AddImage(img)
	if err != nil {
		return err
	}
	return nil
}


func NewController(root string, domain string, redisPort string, redisAddr string) *Controller {
	return &Controller{WebRoot: root, Cache: helpers.NewCache(),
								Domain: domain, RedisConfig: helpers.RedisConf{
																		Port: redisPort,
																		Addr: redisAddr,
																		},
																	}
}
