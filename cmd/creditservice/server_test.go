package main

import "testing"

func TestCreateApplicationRequestValidate(t *testing.T) {
	/** 
	WHY: table-driven
		Instead of N separate test functions, we compile all the cases into a table 
		and then run each one through the same check in a single loop
	*/
	tests := []struct {
		name string // What is the name of the test
		req createApplicationRequest // What we are passing on
		wantErr bool // Are we expecting an error or not
	}{
		{
			name: "valid request",
			req: createApplicationRequest{Client: "Alina", Amount: 1000, Term: 12},
			wantErr: false,
		},
		{
			name: "empty client",
			req: createApplicationRequest{Client: "", Amount: 1000, Term: 12},
			wantErr: true,
		},
		{
			name: "whitespace client",
			req: createApplicationRequest{Client: "   ", Amount: 1000, Term: 12},
			wantErr: true,
		},
		{
			name: "zero amount",
			req: createApplicationRequest{Client: "Alina", Amount: 0, Term: 12},
			wantErr: true,
		},
		{
			name: "negative amount",
			req: createApplicationRequest{Client: "Alina", Amount: -2, Term: 12},
			wantErr: true,
		},
		{
			name: "zero term",
			req: createApplicationRequest{Client: "Alina", Amount: 1000, Term: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		// WHY: t.Run needs a function to execute for each test case
		// NOTE: We use an anonymous function because we only need this code here inside the loop
		// It automatically "sees" and uses the current 'tt' variable from the loop
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()

			if tt.wantErr && err == nil {
				t.Errorf("expected an error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}