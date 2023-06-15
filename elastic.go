package elastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

func request(method Method, endpoint string, params string, waitToRefresh ...bool) (interface{}, error) {
	if !IsInitiated() {
        return nil, errors.New("elastic lib is not initiated")
    }

    if !strings.HasPrefix(endpoint, "/") {
        endpoint = "/" + endpoint
    }

    url := elasticUrl + endpoint
    if len(waitToRefresh) > 0 && waitToRefresh[0] {
		url += "?refresh=wait_for"
	}

    if len(params) > 0 {
	    lastQuery = url + "\n" + params
    }

    req, err := http.NewRequest(method.String(), url, strings.NewReader(params))
	if err != nil {
		return nil, err
	}
    req.Header.Add("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var result interface{}
	d := json.NewDecoder(resp.Body)
	// use json.Number instead of float64
	d.UseNumber()
	if err := d.Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func Request(method Method, endpoint string, params string, waitToRefresh ...bool) (map[string]interface{}, error) {
    result, err := request(method, endpoint, params, waitToRefresh...)
    if err != nil {
        return nil, err
    }

    return result.(map[string]interface{}), nil
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

func parseSetItemResp(resp interface{}, action Action) (bool, bool, error) {
	resp, success := resp.(map[string]interface{})[action.String()]
	if !success {
		return false, false, nil
	}

	result, success := resp.(map[string]interface{})["result"]
	if success {
		if result.(string) == action.String()+"d" {
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
		actions := []Action{
			ActionCreate,
		    ActionUpdate,
			ActionDelete,
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

func parseEditItemResponse(result map[string]interface{}, action string) (string, error) {
    entId, entIdOk := result["_id"].(string)
    status, statusOk := result["result"].(string)

    if !entIdOk || !statusOk || status != action + "d" {
        if !statusOk {
            status, statusOk = result["error"].(string); if !statusOk {
                status = "unknown"
            }
        }

        return entId, errors.New(fmt.Sprintf("Failed to %s elastic entity: %v", action, status))
    }

    return entId, nil
}

func CatIndices(target ...string) ([]Indice, error) {
    var indices []Indice

    if !IsInitiated() {
        return indices, errors.New("Elastic is not initiated")
    }

    endpoint := "/_cat/indices"
    if len(target) > 0 {
        endpoint += "/" + target[0]
    }
    endpoint += "?format=json"

    result, err := request(MethodGet, endpoint, "")
    if err != nil {
        return indices, err
    }

    for _, i := range result.([]interface{}) {
        item := i.(map[string]interface{})
        index, ok := item["index"]; if !ok {
            return indices, errors.New(fmt.Sprintf("No index found in cat indices response: %v", item))
        }

        prCnt, err := strconv.Atoi(item["pri"].(string)); if err != nil {
            return indices, err
        }
        rCtn, err := strconv.Atoi(item["rep"].(string)); if err != nil {
            return indices, err
        }
        dCnt, err := strconv.Atoi(item["docs.count"].(string)); if err != nil {
            return indices, err
        }
        ddCnt, err := strconv.Atoi(item["docs.deleted"].(string)); if err != nil {
            return indices, err
        }

        indices = append(indices, Indice{
            Index: index.(string),
            Health: item["health"].(string),
            Status: item["status"].(string),
            Uuid: item["uuid"].(string),
            PrimariesCnt: prCnt, 
            ReplicasCnt: rCtn, 
            DocsCnt: dCnt, 
            DocsDeletedCnt: ddCnt,
            StoreSize: item["store.size"].(string),
            PrimaryStoreSize: item["pri.store.size"].(string),
        })
    }

    return indices, nil
}

func parseError(result map[string]interface{}) error {
    elErr, ok := result["error"]; if !ok {
        return nil
    }

    return errors.New(fmt.Sprintf(
        "[Elastic error] %v: %v", 
        elErr.(map[string]interface{})["type"],
        elErr.(map[string]interface{})["reason"],
    ))
}

func Docs() *Doc {
    if docs == nil {
        docs = &doc{}
    }

    return docs
}

func Indexes() *Index {
    if indexes == nil {
        indexes = &index{}
    }

    return indexes
}
