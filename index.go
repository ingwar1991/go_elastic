package elastic

import (
	"errors"
	"fmt"
	"strconv"
)

type Index interface {
    Get(indexName string, params map[string]interface{}) (map[string]interface{}, error)

    Exists(indexName string, params map[string]interface{}) ([]interface{}, error)

    Create(indexStruct IndexStructure, waitForActiveShards ...int) error
    
    Delete(indexName string) error
    
    GetMapping(indexName string) (map[string]interface{}, error)
    
    UpdateMapping(indexName string, props map[string]interface{}) error
}

type index struct {}

func (i *index) Get(indexName string) (IndexStructure, error) {
    var indexStructure IndexStructure

    result, err := Request(MethodGet, "/" + indexName, "")
    if err != nil {
        return indexStructure, err
    }
    
    result, ok := result[ indexName ].(map[string]interface{}); if !ok {
        return indexStructure, errors.New(fmt.Sprintf("Unknown error at index.Get: %v", result))
    }

    indexStructure.Name = result[ indexName ].(string)
    indexStructure.Aliases = result["aliases"].(map[string]interface{})
    indexStructure.Mappings = result["mappings"].(map[string]interface{})
    indexStructure.Settings = result["settings"].(map[string]interface{})

    return indexStructure, nil
}

func (i *index) Exists(indexName string) (bool, error) {
    result, err := request(MethodHead, "/"+indexName, "")
    if err != nil {
        return false, err 
    }

    return result.(string) == "200 - true", nil
}

func (i *index) Create(indexStruct IndexStructure, waitForActiveShards ...int) error {
    entJson, err := toJson(indexStruct)
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
    }

    endpoint := "/"+indexStruct.Name
    if len(waitForActiveShards) > 0 {
        endpoint += "?wait_for_active_shards="+strconv.Itoa(waitForActiveShards[0])
    }

    result, err := Request(MethodPut, endpoint, entJson)
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to create elastic index: %v", err))
    }

    indexName, ok := result["index"].(string)
    if !ok || indexName != indexStruct.Name {
        err := parseError(result)
        if err == nil {
            err = errors.New(fmt.Sprintf("Unknown error at index.Create: %v", result))
        }

        return err
    }

    return nil
}

func (i *index) Delete(indexName string) error {
    result, err := Request(MethodDelete, "/"+indexName, "")
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to delete elastic index: %v", err))
    }

    err = nil
    if !result["acknowledged"].(bool) {
        err = parseError(result)
        if err == nil {
            err = errors.New(fmt.Sprintf("Unknown error at index.Delete: %v", result))
        }
    }

    return err 
}
    
func (i *index) GetMapping(indexName string, fieldName string) (map[string]interface{}, error) {
    endpoint := "/"+indexName+"/_mapping"
    if fieldName != "" {
        endpoint += "/field/"+fieldName
    }

    result, err := Request(MethodGet, endpoint, "")
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Failed to get elastic mapping: %v", err))
    }

    result, ok := result[ indexName ].(map[string]interface{})
    if !ok {
        err = parseError(result)
        if err == nil {
            err = errors.New(fmt.Sprintf("Unknown error at index.GetMapping: %v", result))
        }

        return nil, err
    }

    result, ok = result["mappings"].(map[string]interface{}); if !ok {
        return nil, errors.New(fmt.Sprintf("Unknown error at index.GetMapping: %v", result))
    }

    if fieldName != "" {
        result, ok = result[ fieldName ].(map[string]interface{}); if !ok {
            return nil, errors.New(fmt.Sprintf("Unknown error at index.GetMapping: %v", result))
        }
    }

    return result, nil
}

func (i *index) UpdateMapping(indexName string, mappings map[string]interface{}) error {
    _, ok := mappings["properties"]; if !ok {
        return errors.New(fmt.Sprintf("Not found `properties` key in new mappings data at index.UpdateMapping: %v", mappings))
    }

    mappingsJson, err := toJson(mappings)
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to json elastic mappings: %v", err))
    }

    result, err := Request(MethodPut, "/"+indexName+"/_mapping", mappingsJson)
    if err != nil {
        err = parseError(result)
        if err == nil {
            err = errors.New(fmt.Sprintf("Unknown error at index.UpdateMapping: %v", result))
        }

        return err
    }

    aknowledged, ok := result["acknowledged"].(bool); if !ok || !aknowledged {
        err = parseError(result)
        if err == nil {
            err = errors.New(fmt.Sprintf("Unknown error at index.UpdateMapping: %v", result))
        }

        return err
    }

    return nil
}
