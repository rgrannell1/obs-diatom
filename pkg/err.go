package diatom

const (
	ERR_JSON_TO_MARKDOWN = "ERR_JSON_TO_MARKDOWN"
)

type CodedError struct {
	Code string
	Err  error
}

func (err *CodedError) Error() string {
	return ERR_JSON_TO_MARKDOWN + ": " + err.Err.Error()
}
