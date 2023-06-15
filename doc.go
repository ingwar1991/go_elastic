package elastic

import (
    "fmt"
    "errors"
)

type Doc interface {
    Get(entityId string, indexName string) (map[string]interface{}, error)

    MGet(entityIds []string, indexName string) ([]map[string]interface{}, error)

    Search(query map[string]interface{}, indexName string) ([]interface{}, int, error)

    Create(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)
    
    Update(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)
    
    Delete(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)
    
    Set(entities SetParams, indexName string, waitToRefresh ...bool) SetResult
}

type doc struct {}

func (i *doc) Get(entityId string, indexName string) (map[string]interface{}, error) {
    if len(entityId) == 0 {
        return nil, errors.New("No entity id transmitted")
    }

    res, err := Request(MethodGet, "/"+indexName+"/_doc/"+entityId)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Failed to get elastic entity: %v", err))
    }

    errStr, ok := res["error"].(string); if ok {
        return nil, errors.New(errStr)
    }

    found, ok := res["found"].(bool); if !ok {
        return nil, errors.New(fmt.Sprintf("Unknown error at doc.Get: %v", res))
    }

    if !found {
        return nil, nil
    }

    return res["_source"].(map[string]interface{}), nil
}

func (i *doc) MGet(entityIds []string, indexName string) ([]map[string]interface{}, error) {
    lenEntityIds := len(entityIds)
    for i := lenEntityIds - 1; i >= 0; i-- {
        if len(entityId) == 0 {
            entityIds = append(entityIds[:i], entityIds[i+1:]...)
        }
    }
    if len(entityIds) == 0 {
        return nil, errors.New("No entity ids transmitted")
    }

    params, err := toJson([string]string{}{
        "ids": entityIds,
    })
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Failed to json elastic entity ids: %v", err))
    }

    res, err := Request(MethodGet, "/"+indexName+"/_mget", params)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Failed to get elastic entities: %v", err))
    }

    elErr := parseError(res); if elErr != nil {
        return nil, elErr
    }

    docs, ok := res["docs"].(map[string]interface{}); if !ok {
        return nil, errors.New(fmt.Sprintf("Unknown error at doc.MGet: %v", res))
    }

    found, ok := docs["found"].(bool); if !ok {
        return nil, errors.New(fmt.Sprintf("Unknown error at doc.Get: %v", res))
    }
    res = nil

    if !found {
        return make([]map[string]interface{}), nil
    }

    return res["_source"].([]map[string]interface{}), nil
}

func (i *doc) Create(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error) {
    action := "create"

    delete(entity, "_id")
    var entId string

    entJson, err := toJson(entity)
    if err != nil {
        return entId, errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
    }

    endpoint := "/"+indexName+"/_doc"
    result, err := Request(MethodPost, endpoint, entJson, waitToRefresh...)
    if err != nil {
        return entId, errors.New(fmt.Sprintf("Failed to %s elastic entity: %v", action, err))
    }

    return parseEditItemResponse(result, action)
}

func (i *doc) Update(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error) {
    action := "update"

    entId, ok := entity["_id"].(string); if !ok {
        return "", errors.New(fmt.Sprintf("No _id transmitted for %s stmt: %v", action, entity))
    }
    delete(entity, "_id")

    entJson, err := toJson(entity)
    if err != nil {
        return entId, errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
    }

    endpoint := "/"+indexName+"/_doc/"+entId
    result, err := Request(MethodPut, endpoint, entJson, waitToRefresh...)
    if err != nil {
        return entId, errors.New(fmt.Sprintf("Failed to %s elastic entity: %v", action, err))
    }

    return parseEditItemResponse(result, action)
}

func (i *doc) Delete(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error) {
    action := "delete"

    entId, ok := entity["_id"].(string); if !ok {
        return "", errors.New(fmt.Sprintf("No _id transmitted for %s stmt: %v", action, entity))
    }
    delete(entity, "_id")

    entJson, err := toJson(entity)
    if err != nil {
        return entId, errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
    }

    endpoint := "/"+indexName+"/_doc/"+entId
    result, err := Request(MethodDelete, endpoint, entJson, waitToRefresh...)
    if err != nil {
        return entId, errors.New(fmt.Sprintf("Failed to %s elastic entity: %v", action, err))
    }

    return parseEditItemResponse(result, action)
}

func (i *doc) Search(query map[string]interface{}, indexName string) ([]interface{}, int, error) {
    var entities []interface{}
    var totalFound int

    if !IsInitiated() {
        return entities, totalFound, errors.New("Elastic is not initiated")
    }

	queryJson, err := toJson(query)
	if err != nil {
		return entities, totalFound, err
	}

    endpoint := "/"+indexName+"/_search"
	result, err := Request(MethodGet, endpoint, queryJson)
	if err != nil {
		return entities, totalFound, err
	}

    elErr := parseError(result); if elErr != nil {
        return entities, totalFound, elErr
    }

    hits, ok := result["hits"].(map[string]interface{}); if !ok {
        return entities, totalFound, errors.New("Failed to parse elastic result")
    }

    entities = hits["hits"].([]interface{})
    totalFound = hits["total"].(map[string]interface{})["value"].(int)

    return entities, totalFound, nil
}

func (i *doc) Set(entities SetParams, indexName string, waitToRefresh ...bool) SetResult {
	addStmts, err := getAddStmts(entities.ToAdd, indexName)
	if err != nil {
		return SetResult{0, 0, 0, 0, []error{err}}
	}

	updateStmts, err := getUpdateStmts(entities.ToUpdate, indexName)
	if err != nil {
		return SetResult{0, 0, 0, 0, []error{err}}
	}

    deleteStmts, err := getDeleteStmts(entities.ToDelete, indexName)
	if err != nil {
        return SetResult{0, 0, 0, 0, []error{err}}
	}
    
	endpoint := "/_bulk"
    result, err := Request(MethodPost, endpoint, addStmts + updateStmts + deleteStmts, waitToRefresh...)
	if err != nil {
        return SetResult{0, 0, 0, 0, []error{err}}
	}

	return parseSetResponse(result)
}
