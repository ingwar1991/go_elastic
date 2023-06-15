package elastic


const (
    ActionCreate Action = "create"
    ActionUpdate Action = "update"
    ActionDelete Action = "delete"
)

const (
    MethodHead Method = "HEAD"
    MethodGet Method = "GET"
    MethodPost Method = "POST"
    MethodPut Method = "PUT"
    MethodDelete Method = "DELETE"
)

const DateFormatElastic = "2006-01-02T15:04:05"
const DateFormat = "2006-01-02 15:04:05"

var elasticConfig Config
var elasticUrl string

var lastQuery string

var docs *doc
var indexes *index

