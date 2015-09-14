package himawari

// Request is a representation of a JSON request for upload sent by
// the user.
type Request struct {
	Filename string
	Mime     string
	Length   int
}

// PendingUpload is the information needed to process a file upload based
// on an accepted request.
// It is also used to generate the JSON response to the user.
type PendingUpload struct {
	Url      string `json:"url"`
	Timeout  int    `json:"timeout"`
	Filename string `json:"filename"`
	Length   int    `json:"-"`
}
