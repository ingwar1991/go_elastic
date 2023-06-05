package elastic

import (
	"errors"
	"fmt"
	"strings"
	"time"
    "encoding/json"
    "net/http"
)

type Config struct {
    Host string
    Port int
    User string
    Password string
}

type SetParams struct {
    ToAdd []map[string]interface{}
    ToUpdate []map[string]interface{}
    ToDelete []map[string]interface{}
}

type SetResult struct {
    Added   int
    Updated int
    Deleted int
    Failed  int
    Errors  []error
}

var elasticConfig Config
var elasticUrl string

var lastQuery string

const DateFormatElastic = "2006-01-02T15:04:05"
const DateFormat = "2006-01-02 15:04:05"

func Init(config Config) error {
    if config.Host == "" {
        return errors.New("Elastic host is not set")
    }
    elasticConfig = config

    patternStr := "%s"
    var patternsArgs []interface{}

    if elasticConfig.User != "" {
        patternsArgs = append(patternsArgs, elasticConfig.User)

        if elasticConfig.Password != "" {
            patternStr = "%s:%s@" + patternStr
            patternsArgs = append(patternsArgs, elasticConfig.Password)
        } else {
            patternStr = "%s@" + patternStr
        }
    }

    patternsArgs = append(patternsArgs, elasticConfig.Host)

    if elasticConfig.Port != 0 {
        patternStr = patternStr + ":%d"
        patternsArgs = append(patternsArgs, elasticConfig.Port)
    }

    patternStr = "https://" + patternStr

    elasticUrl = fmt.Sprintf(patternStr, patternsArgs...)

    return nil
}

func IsInitiated() bool {
    return elasticUrl != ""
}

func toJson(entity interface{}) (string, error) {
	res, err := json.Marshal(entity)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func requestPostJson(url string, params string) (map[string]interface{}, error) {
	lastQuery = url + "\n" + params

    resp, err := http.Post(url, "application/json", strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	d := json.NewDecoder(resp.Body)
	// use json.Number instead of float64
	d.UseNumber()
	if err := d.Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func FromElasticDate(date string) (string, error) {
	// fixing wrong dates format
	date = strings.Split(date, ".")[0]
	date = strings.Split(date, "+")[0]

	dateParsed, err := time.Parse(DateFormatElastic, date)
	if err != nil {
		return "", err
	}

	return dateParsed.Format(DateFormat), nil
}

func ToElasticDate(date string) (string, error) {
	dateParsed, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", err
	}

	return dateParsed.Format(DateFormatElastic), nil
}

func getAddStmts(entities []map[string]interface{}, indexName string) (string, error) {
	if len(entities) < 1 {
		return "", nil
	}

	createStmt := "{\"create\":{\"_index\":\"" + indexName + "\"}}"

	stmts := ""
	for _, entity := range entities {
		delete(entity, "_id")

		entJson, err := toJson(entity)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
		}

		stmts += createStmt + "\n" + entJson + "\n"
	}

	return stmts, nil
}

func getUpdateStmts(entities []map[string]interface{}, indexName string) (string, error) {
	if len(entities) < 1 {
		return "", nil
	}

	stmts := ""
	for _, entity := range entities {
        id, ok := entity["_id"].(string); if !ok {
            return "", errors.New(fmt.Sprintf("No _id transmitted for update stmts: %v", entity))
        }

		updateStmt := "{\"update\":{\"_index\":\"" + indexName + "\",\"_id\": \"" + id + "\"}}"

		delete(entity, "_id")
		entJson, err := toJson(entity)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
		}

		stmts += updateStmt + "\n{\"doc\":" + entJson + "}\n"
	}

	return stmts, nil
}

func getDeleteStmts(entities []map[string]interface{}, indexName string) (string, error) {
	if len(entities) < 1 {
		return "", nil
	}

	stmts := ""
	for _, entity := range entities {
        id, ok := entity["_id"].(string); if !ok {
            return "", errors.New(fmt.Sprintf("No _id transmitted for delete stmts: %v", entity))
        }

		stmts += "{\"delete\":{\"_index\":\"" + indexName + "\",\"_id\": \"" + id + "\"}}\n"
	}

	return stmts, nil
}


func parseSetItemResp(resp interface{}, action string) (bool, bool, error) {
	resp, success := resp.(map[string]interface{})[action]
	if !success {
		return false, false, nil
	}

	result, success := resp.(map[string]interface{})["result"]
	if success {
		if result.(string) == action+"d" {
			return true, true, nil
		}

		return true, false, nil
	}

	resp, success = resp.(map[string]interface{})["error"]
	if !success {
		return true, false, nil
	}

	resp, success = resp.(map[string]interface{})["reason"]
	if !success {
		return true, false, nil
	}

	return true, false, errors.New(resp.(string))
}

func parseSetResponse(result map[string]interface{}) SetResult {
	res := SetResult{}

	items := result["items"].([]interface{})
	for _, item := range items {
		actions := []string{
			"create",
			"update",
			"delete",
		}

		foundAction := false
		for _, action := range actions {
			isAction, success, err := parseSetItemResp(item, action)
			if isAction {
				foundAction = true

				if success {
					switch action {
					case "create":
						res.Added++
						break
					case "update":
						res.Updated++
						break
					case "delete":
					default:
						res.Deleted++
						break
					}
				} else {
					res.Failed++
				}

				if err != nil {
					res.Errors = append(res.Errors, err)
				}

				continue
			}
		}

		if !foundAction {
			res.Failed++

			jsonItem, err := toJson(item)
			if err != nil {
				res.Errors = append(res.Errors, errors.New("no action: " + err.Error()))
			} else {
				res.Errors = append(res.Errors, errors.New("no action: " + jsonItem))
			}
		}
    }

	return res
}

func Search(query map[string]interface{}, indexName string) ([]interface{}, int, error) {
    var entities []interface{}
    var totalFound int

    if !IsInitiated() {
        return entities, totalFound, errors.New("Elastic is not initiated")
    }

	queryJson, err := toJson(query)
	if err != nil {
		return entities, totalFound, err
	}

	result, err := requestPostJson(elasticUrl+"/"+indexName+"/_search", queryJson)
	if err != nil {
		return entities, totalFound, err
	}

    elErr, ok := result["error"]; if ok {
        return entities, totalFound, errors.New(fmt.Sprintf(
            "[Elastic error] %v: %v", 
            elErr.(map[string]interface{})["type"],
            elErr.(map[string]interface{})["reason"],
        ))
    }

    hits, ok := result["hits"].(map[string]interface{}); if !ok {
        return entities, totalFound, errors.New("Failed to parse elastic result")
    }

    entities = hits["hits"].([]interface{})
    totalFound = hits["total"].(map[string]interface{})["value"].(int)

    return entities, totalFound, nil
}

func Set(entities SetParams, indexName string, waitToRefresh ...bool) SetResult {
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
    
	apiCallUrl := elasticUrl + "/_bulk"
	if len(waitToRefresh) > 0 && waitToRefresh[0] {
		apiCallUrl += "?refresh=wait_for"
	}

	result, err := requestPostJson(apiCallUrl, addStmts + updateStmts + deleteStmts)
	if err != nil {
        return SetResult{0, 0, 0, 0, []error{err}}
	}

	return parseSetResponse(result)
}

func Add(entities []map[string]interface{}, indexName string, waitToRefresh ...bool) (int, int, []error) {
    result := Set(SetParams{ToAdd: entities}, indexName, waitToRefresh...)

    return result.Added, result.Failed, result.Errors
}

func Update(entities []map[string]interface{}, indexName string, waitToRefresh ...bool) (int, int, []error) {
    result := Set(SetParams{ToUpdate: entities}, indexName, waitToRefresh...)

    return result.Updated, result.Failed, result.Errors
}

func Delete(entities []map[string]interface{}, indexName string, waitToRefresh ...bool) (int, int, []error) {
    result := Set(SetParams{ToDelete: entities}, indexName, waitToRefresh...)

    return result.Deleted, result.Failed, result.Errors
}
