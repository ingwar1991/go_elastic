package elastic

import (
    "testing"
    "strings"
)

func TestInitFull(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
        User: "user",
        Password: "password",
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != "https://user:password@localhost:9200" {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestInitWithoutUser(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != "https://localhost:9200" {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestInitWithoutPassword(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != "https://localhost:9200" {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestInitWithoutPort(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        User: "user",
        Password: "password",
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != "https://user:password@localhost" {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestSet(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    var entitiesToUpdate []map[string]interface{}
    entitiesToUpdate = append(entitiesToUpdate, map[string]interface{}{"_id": "1", "Name": "name 1", "City": "city 1"})
    entitiesToUpdate = append(entitiesToUpdate, map[string]interface{}{"_id": "2", "Name": "name 2"})
    
    var entitiesToAdd []map[string]interface{}
    entitiesToAdd = append(entitiesToAdd, map[string]interface{}{"Name": "name 3"})
    entitiesToAdd = append(entitiesToAdd, map[string]interface{}{"Name": "name 4", "City": "city 4"})
   
    Set(entitiesToAdd, entitiesToUpdate, "test", false)
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to set: %v", lastQuery)
    }
}

func TestSearch(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    Search(map[string]interface{}{
        "query": map[string]interface{}{
            "match_all": map[string]string{
                "name": "name 1",
            },
        },
        "size": 10,
    }, "test")

    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/test/_search" {
        t.Errorf("Failed to search: %v", lastQuery)
    }
}
