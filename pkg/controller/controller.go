package controller

import (
	"fmt"
	"log"

	"adeptus-mechanicus.void/git/keiji/pkg/helpers"
)

type Controller struct{
	WebRoot		string
	Domain		string
	RedisConfig helpers.RedisConf
	Cache		*helpers.AllCache
}

func (c *Controller) Headers() *helpers.HeaderCollection {
	headers, err := helpers.GetHeaders(c.RedisConfig)
	if err != nil {
		log.Fatal(err, "Headers couldnt be loaded. Exiting.")
	}
	return headers
}

func (c *Controller) Menu() *helpers.MenuElement {
	links, err := helpers.GetMenuLinks(c.RedisConfig)
	if err != nil {
		log.Fatal(err, "Menu links couldnt be couldnt be loaded. Exiting.")
	}
	return links
}

func (c *Controller) AdminTables() *helpers.AdminTables {
	tables, err := helpers.GetAdminTables(c.RedisConfig)
	if err != nil {
		log.Fatal(err, "Administrator tables couldnt be couldnt be loaded. Exiting.")
	}
	return tables
}

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

func NewController(root string, domain string, redisPort string, redisAddr string) *Controller {
	return &Controller{WebRoot: root, Cache: helpers.NewCache(),
								Domain: domain, RedisConfig: helpers.RedisConf{
																		Port: redisPort,
																		Addr: redisAddr,
																		},
																	}
}
