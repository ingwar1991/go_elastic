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
		updateStmt := "{\"update\":{\"_index\":\"" + indexName + "\",\"_id\": \"" + entity["_id"].(string) + "\"}}"

		delete(entity, "_id")
		entJson, err := toJson(entity)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to json elastic entity: %v", err))
		}

		stmts += updateStmt + "\n{\"doc\":" + entJson + "}\n"
	}

	return stmts, nil
}

type SetResult struct {
	Added   int
	Updated int
	Deleted int
	Failed  int
	Errors  []string
}

/*
SUCCESS examples:
{"create":{...,"result":"created",...}},
{"update":{...,"result":"updated",...}}
ERROR examples:
{"create":{...,"error":{...,"reason":STRING,...}}},
{"update":{...,"error":{...,"reason":STRING,...}}},
*/
func parseSetItemResp(resp interface{}, action string) (bool, bool, string) {
	resp, success := resp.(map[string]interface{})[action]
	if !success {
		return false, false, ""
	}

	result, success := resp.(map[string]interface{})["result"]
	if success {
		if result.(string) == action+"d" {
			return true, true, ""
		}

		return true, false, ""
	}

	resp, success = resp.(map[string]interface{})["error"]
	if !success {
		return true, false, ""
	}

	resp, success = resp.(map[string]interface{})["reason"]
	if !success {
		return true, false, ""
	}

	return true, false, resp.(string)
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

				if err != "" {
					res.Errors = append(res.Errors, err)
				}

				continue
			}
		}

		if !foundAction {
			res.Failed++

			jsonItem, err := toJson(item)
			if err != nil {
				res.Errors = append(res.Errors, "no action: "+err.Error())
			} else {
				res.Errors = append(res.Errors, "no action: "+jsonItem)
			}
		}
	}

	return res
}

func Search(query map[string]interface{}, indexName string) ([]interface{}, error) {
    if !IsInitiated() {
        return nil, errors.New("Elastic is not initiated")
    }

	queryJson, err := toJson(query)
	if err != nil {
		return nil, err
	}

	result, err := requestPostJson(elasticUrl+"/"+indexName+"/_search", queryJson)
	if err != nil {
		return nil, err
	}

    elErr, ok := result["error"]; if ok {
        return nil, errors.New(fmt.Sprintf(
            "[Elastic error] %v: %v", 
            elErr.(map[string]interface{})["type"],
            elErr.(map[string]interface{})["reason"],
        ))
    }

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

    return hits, nil
}

func Set(entitiesToAdd []map[string]interface{}, entitiesToUpdate []map[string]interface{}, indexName string, waitToRefresh ...bool) SetResult {
	addStmts, err := getAddStmts(entitiesToAdd, indexName)
	if err != nil {
		return SetResult{0, 0, 0, 0, []string{err.Error()}}
	}

	updateStmts, err := getUpdateStmts(entitiesToUpdate, indexName)
	if err != nil {
		return SetResult{0, 0, 0, 0, []string{err.Error()}}
	}

	// TESTING
	//fmt.Println("update stmts: ", updateStmts)
	//fmt.Println("add stmts: ", addStmts)
	//fmt.Println("_BULK stmts: ", addStmts+updateStmts)
	//return SetResult{0, 0, 0, 0, []string{},}

	apiCallUrl := elasticUrl + "/_bulk"
	if len(waitToRefresh) > 0 && waitToRefresh[0] {
		apiCallUrl += "?refresh=wait_for"
	}

	result, err := requestPostJson(apiCallUrl, addStmts+updateStmts)
	if err != nil {
		return SetResult{0, 0, 0, 0, []string{err.Error()}}
	}

	return parseSetResponse(result)
}
