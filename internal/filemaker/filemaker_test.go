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

func TestTargetAuthorizationUsesAllowlists(t *testing.T) {
	target := Target{Name: "primary", Adapter: "dataapi", BaseURL: "https://fm.example.com", Credential: "fm", DefaultDatabase: "CRM", AllowedDatabases: []string{"CRM"}, AllowedLayouts: []string{"Customers"}, AllowedOperations: []string{"find", "update"}, AllowedScripts: []string{"Refresh Cache"}}
	if err := target.Validate(); err != nil {
		t.Fatal(err)
	}
	request := Request{Operation: "find", Layout: "Customers", ScriptAfter: &ScriptRequest{Name: "Refresh Cache"}}
	if err := target.Authorize(&request); err != nil {
		t.Fatal(err)
	}
	if request.Database != "CRM" {
		t.Fatal("default database was not applied")
	}
	bad := Request{Operation: "find", Database: "Other", Layout: "Customers"}
	if err := target.Authorize(&bad); err == nil {
		t.Fatal("expected database allow-list rejection")
	}
}
