# go_elastic
Library for quering elasticSearch

## Usage

### Import
First, import the library

```
import (
    elastic "github.com/ingwar1991/go_elastic"
)
```

### Init
You will need to call Init() function with Config{}

**Host** - **_required_**

**Port** - optional

**User** - optional( **_required_** if **Pass** was transmitted )

**Pass** - optional

```
err := elastic.Init(Config{
    Host: "localhost",
    Port: 9200,
    User: "user",
    Password: "password",
})
if err != nil {
    fmt.Errorf("Failed to init: %v", err)
}
```

### Choose docs or indexes
The library has 2 entities with unique methods:
* Docs
* Indexes

```
elastic.Docs()
elastic.Indexes()
```

#### Docs methods

```
Get(entityId string, indexName string) (map[string]interface{}, error)

MGet(entityIds []string, indexName string) ([]map[string]interface{}, error)

Search(query map[string]interface{}, indexName string) ([]interface{}, int, error)

Create(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)

Update(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)

Delete(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)

Set(entities SetParams, indexName string, waitToRefresh ...bool) SetResult
```

##### Get
Getting entity by entityId( _id )

`func Get(entityId string, indexName string) (map[string]interface{}, error)`

##### MGet
Gets multiple entities by ids( _id )

`func MGet(entityIds []string, indexName string) ([]map[string]interface{}, error)`

##### Searching
`func Search(query map[string]interface{}, indexName string) ([]interface{}, int, error)`

Search function returns list of "hits" from the elastic response, a number of found entities and an error 

```
entities, totalFound, err := elastic.Search(map[string]interface{}{
    "query": map[string]interface{}{
        "match_all": map[string]string{
            "name": "name 1",
        },
    },
    "size": 10,
}, "test")
if err != nil {
    fmt.Errorf("Failed to query search query: %v", err)
}
```

#### Creating/Updating/Deleting
Library has both separated Create/Update/Delete methods and combined Set method

All methods get last **_optional_** parameter `waitToRefresh` for waiting for elastic to refresh the index data
> "?refresh=wait_for"

All but `Set` methods expect entity as first parameter and return entityId( `_id` ) and optionally error

##### Create
`func Create(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)`

##### Update 
`func Update(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)`

##### Delete
`func Delete(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, []error)`

##### Set
`func Set(entities SetParams, indexName string, waitToRefresh ...bool) SetResult`

Set function return SetResult
```
type SetResult struct {
	Added   int
	Updated int
	Deleted int
	Failed  int
	Errors  []string
}
```

```
entities := elastic.SetParams{
    []map[string]interface{}{
        {"Name": "name 3"},
        {"Name": "name 4", "City": "city 4"},
    },
    []map[string]interface{}{
        {"_id": "1", "Name": "name 1", "City": "city 1"},
        {"_id": "2", "Name": "name 2"},
    },
    []map[string]interface{}{
        {"_id": "5", "Name": "name 5", "City": "city 5"},
        {"_id": "6"},
    },
}

result := elastic.Set(entities, "test")

if len(result.Errors) > 0 {
   for _, err := range result.Errors {
        fmt.Errorf("Error from elastic.Set(): %v", err)
   }
}

fmt.Sprintf("Result[ added: %v, updated: %v, deleted: %v, failed: %v ]", 
    result.Added,
    result.Updated,
    result.Deleted,
    result.Failed,
)
```

#### Indexes methods
```
Get(indexName string, params map[string]interface{}) (map[string]interface{}, error)

Exists(indexName string, params map[string]interface{}) ([]interface{}, error)

Create(indexStruct IndexStructure, waitForActiveShards ...int) error

Delete(indexName string) error

GetMapping(indexName string) (map[string]interface{}, error)

UpdateMapping(indexName string, props map[string]interface{}) error
```
