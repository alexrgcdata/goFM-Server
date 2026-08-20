package filemaker

import "testing"

func TestRequestValidation(t *testing.T) {
	valid := Request{Operation: "find", Database: "CRM", Layout: "Customers", Limit: 100}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []Request{{Operation: "script", Database: "CRM", Layout: "Customers"}, {Operation: "find", Database: "", Layout: "Customers"}, {Operation: "find", Database: "CRM", Layout: "Customers", Limit: 1001}, {Operation: "find", Database: "CRM", Layout: "Customers", Offset: -1}} {
		if err := invalid.Validate(); err == nil {
			t.Fatalf("expected validation error for %#v", invalid)
		}
	}
}
