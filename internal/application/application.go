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

// Store represents a simple memory cell
type Store struct {
	items []Application
	nextID int
}

// NewStore create and returns an exemplar of Store
func NewStore() *Store {
	return &Store{nextID: 1}
}

// Add accept the request and saves it to memory, returning the completed request
func (s *Store) Add(app Application) Application { // WHY: `*` is used so that the original is modified, rather than a copy
	app.ID = s.nextID
	app.Status = "new"

	s.nextID++ // NOTE: shift the counter
	s.items = append(s.items, app) // NOTE: place in storage

	return app
}