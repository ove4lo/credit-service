package application

// Application represents an individual request
type Application struct {
	// NOTE: json:"..." - what is the field called in json
	ID int `json:"id"` 
	Client string `json:"client"` // WHY: with a capital, so that the field is visible from outside the package
	Amount int `json:"amount"` // WHY: without the tag, the JSON would be Amount with a capital
	Term int `json:"term"`
	Status string `json:"status"`
}
