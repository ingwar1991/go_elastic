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

func TestSearch(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    entities, totalFound, err := Search(map[string]interface{}{
        "query": map[string]interface{}{
            "match_all": map[string]string{
                "name": "name 1",
            },
        },
        "size": 10,
    }, "test")

    if len(entities) != totalFound {
        t.Errorf("Failed to search: %v", err)
    }

    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/test/_search" {
        t.Errorf("Failed to search: %v", lastQuery)
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
   
    Set(entities, "test")
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to set: %v", lastQuery)
    }
}

func TestAdd(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }
    
    entities := []map[string]interface{}{
        {"Name": "name 3"},
        {"Name": "name 4", "City": "city 4"},
    }
   
    Add(entities, "test")
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to add: %v", lastQuery)
    }
}

func TestUpdate(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }
    
    entities := []map[string]interface{}{
        {"_id": "1", "Name": "name 1", "City": "city 1"},
        {"_id": "2", "Name": "name 2"},
    }
   
    Update(entities, "test")
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to update: %v", lastQuery)
    }
}

func TestDelete(t *testing.T) {
    err := Init(Config{
        Host: "localhost",
        Port: 9200,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }
    
    entities := []map[string]interface{}{
        {"_id": "5", "Name": "name 5", "City": "city 5"},
        {"_id": "6"},
    }
   
    Delete(entities, "test")
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to delete: %v", lastQuery)
    }
}
