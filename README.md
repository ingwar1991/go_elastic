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
`Search(query map[string]interface{}, indexName string) ([]interface{}, error)`

Search function returns list of "hits" from the elastic response or error 

```
values, err := Search(map[string]interface{}{
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

