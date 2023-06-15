package elastic


type Action string
func (a Action) String() string {
    return string(a)
}

type Method string
func (m Method) String() string {
    return string(m)
}

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

type Indice struct {
    Index string
    Health string
    Status string
    Uuid string
    PrimariesCnt int
    ReplicasCnt int
    DocsCnt int
    DocsDeletedCnt int
    StoreSize string
    PrimaryStoreSize string
}

type IndexStructure struct {
    Name string
    Aliases map[string]interface{} `json:"aliases,omitempty"`
    Mappings map[string]interface{} `json:"mappings,omitempty"`
    Settings map[string]interface{} `json:"settings,omitempty"`
}

type item interface {
    Create(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)
    
//    Update(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)
    
    Delete(entity map[string]interface{}, indexName string, waitToRefresh ...bool) (string, error)
}
