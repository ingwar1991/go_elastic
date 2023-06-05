# go_elastic
Library for quering elastic

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
err := Init(Config{
    Host: "localhost",
    Port: 9200,
    User: "user",
    Password: "password",
})
if err != nil {
    fmt.Errorf("Failed to init: %v", err)
}
```

### Searching
`func Search(query map[string]interface{}, indexName string) ([]interface{}, int, error) {}`

Search function returns list of "hits" from the elastic response, a number of found entities and an error 

```
entities, totalFound, err := Search(map[string]interface{}{
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

### Adding/Updating/Deleting
Library has both separated Add/Update/Delete methods and combined Set method

All methods get last **_optional_** parameter `waitToRefresh` for waiting for elastic to refresh the index data
> "?refresh=wait_for"

#### Set
`func Set(entities SetParams, indexName string, waitToRefresh ...bool) SetResult {}`

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
entities := SetParams{
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

result := Set(entities, "test")

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

#### Add
`func Add(entities []map[string]interface{}, indexName string, waitToRefresh ...bool) (int, int, []error) {}`

```
entities := []map[string]interface{}{
    {"Name": "name 3"},
    {"Name": "name 4", "City": "city 4"},
}

added, failed, errs := Add(entities, "test")
if len(errs) > 0 {
    for _, err := range errs {
        fmt.Errorf("Error from elastic.Add(): %v", err)
    }
}

fmt.Sprintf("Added: %v, Failed: %v", added, failed)
```

#### Update 
`func Update(entities []map[string]interface{}, indexName string, waitToRefresh ...bool) (int, int, []error) {}`

```
entities := []map[string]interface{}{
    {"Name": "name 3"},
    {"Name": "name 4", "City": "city 4"},
}

updated, failed, errs := Update(entities, "test")
if len(errs) > 0 {
    for _, err := range errs {
        fmt.Errorf("Error from elastic.Update(): %v", err)
    }
}

fmt.Sprintf("Updated: %v, Failed: %v", updated, failed)
```

#### Delete
`func Delete(entities []map[string]interface{}, indexName string, waitToRefresh ...bool) (int, int, []error) {}`

```
entities := []map[string]interface{}{
    {"Name": "name 3"},
    {"Name": "name 4", "City": "city 4"},
}

deleted, failed, errs := Delete(entities, "test")
if len(errs) > 0 {
    for _, err := range errs {
        fmt.Errorf("Error from elastic.Delete(): %v", err)
    }
}

fmt.Sprintf("Deleted: %v, Failed: %v", deleted, failed)
```
