package elastic

import (
    "testing"
    "strings"
    "fmt"
)

//const varHost = "localhost"
//const varPort = 9200
//const varUser = "user"
//const varPassword = "password"
//const varIndex = "test"

const varHost = "os-b7096ec-newsworthy-a213.aivencloud.com"
const varPort = 19356
const varUser = "avnadmin"
const varPassword = "AVNS_mm8ZgbEL-_ssV4S"
const varIndex = "igor_testing"

func TestInitFull(t *testing.T) {
    err := Init(Config{
        varHost,
        varPort,
        varUser,
        varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != fmt.Sprintf("https://%s:%s@%s:%d", varUser, varPassword, varHost, varPort) {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestInitWithoutUser(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != fmt.Sprintf("https://%s:%d", varHost, varPort) {
        t.Errorf("Failed to init: %v", elasticUrl)
    } 
}

func TestInitWithoutPassword(t *testing.T) {
    err := Init(Config{
        Host: varHost, 
        Port: varPort,
        User: varUser,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != fmt.Sprintf("https://%s@%s:%d", varUser, varHost, varPort) {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestInitWithoutPort(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    if elasticUrl != fmt.Sprintf("https://%s:%s@%s", varUser, varPassword, varHost) {
        t.Errorf("Failed to init: %v", elasticUrl)
    }
}

func TestCatIndices(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    indices, err := CatIndices()
    url := fmt.Sprintf("https://%s:%s@%s:%d/_cat/indices?format=json", varUser, varPassword, varHost, varPort)
    if strings.Split(lastQuery, "\n")[0] != url {
        t.Errorf("Failed to get indices: %v", lastQuery)
    }

    fmt.Println("Cat Indices: ", indices, err)
}

func TestCatIndicesWithTarget(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    indices, err := CatIndices(varIndex)
    url := fmt.Sprintf("https://%s:%s@%s:%d/_cat/indices/%s?format=json", varUser, varPassword, varHost, varPort, varIndex)
    if strings.Split(lastQuery, "\n")[0] != url {
        t.Errorf("Failed to get indices: %v", lastQuery)
    }

    fmt.Println("Cat Target indice: ", indices, err)
}

func TestDocsSearch(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    entities, totalFound, err := Docs().Search(map[string]interface{}{
        "query": map[string]interface{}{
            "match_all": map[string]string{
                "name": "name 1",
            },
        },
        "size": 10,
    }, varIndex)

    if len(entities) != totalFound {
        t.Errorf("Failed to search: %v", err)
    }

    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/test/_search" {
        t.Errorf("Failed to search: %v", lastQuery)
    }

    fmt.Println("Docs Search: ", entities, totalFound, err)
}

func TestDocsCreate(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    entity := map[string]interface{}{
        "Name": "name 4", 
        "City": "city 4",
    }

    res, err := Docs().Create(entity, varIndex) 
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to create: %v", lastQuery)
    }

    fmt.Println("Docs Create: ", res, err)
}

func TestDocsUpdate(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    entity := map[string]interface{}{
        "_id": "1", 
        "Name": "name 1", 
        "City": "city 1",
    }

    res, err := Docs().Update(entity, varIndex)
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to update: %v", lastQuery)
    }

    fmt.Println("Docs Update: ", res, err)
}

func TestDocsDelete(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
    })
    if err != nil {
        t.Errorf("Failed to init: %v", err)
    }

    entity := map[string]interface{}{
        "_id": "5", 
        "Name": "name 5", 
        "City": "city 5",
    }

    res, err := Docs().Delete(entity, varIndex)
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to delete: %v", lastQuery)
    }

    fmt.Println("Docs Delete: ", res, err)
}

func TestDocsSet(t *testing.T) {
    err := Init(Config{
        Host: varHost,
        Port: varPort,
        User: varUser,
        Password: varPassword,
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
   
    res := Docs().Set(entities, varIndex) 
    if strings.Split(lastQuery, "\n")[0] != "https://localhost:9200/_bulk" {
        t.Errorf("Failed to set: %v", lastQuery)
    }

    fmt.Println("Docs Set: ", res)
}
