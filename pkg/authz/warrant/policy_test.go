package authz

import (
	"testing"
	"time"
)

func TestExpiresIn(t *testing.T) {
	warrant := Warrant{
		ObjectType:  "role",
		ObjectId:    "admin",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-a",
		Policy:      Policy("expiresIn(\"10ms\")"),
		CreatedAt:   time.Now(),
	}

	match, err := warrant.Policy.Eval(PolicyContext{"warrant": &warrant})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	match, err = warrant.Policy.Eval(PolicyContext{"warrant": &warrant})
	if match || err != nil {
		t.Fatalf("Expected match to be false, but it was true, %v", err)
	}
}

func TestGt(t *testing.T) {
	warrant := Warrant{
		ObjectType:  "role",
		ObjectId:    "admin",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-a",
		Policy:      Policy("amount > 50"),
		CreatedAt:   time.Now(),
	}

	match, err := warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  51,
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if match || err != nil {
		t.Fatalf("Expected match to be false, but it was true, %v", err)
	}
}

func TestGte(t *testing.T) {
	warrant := Warrant{
		ObjectType:  "role",
		ObjectId:    "admin",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-a",
		Policy:      Policy("amount >= 50"),
		CreatedAt:   time.Now(),
	}

	match, err := warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  51,
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  49,
	})
	if match || err != nil {
		t.Fatalf("Expected match to be false, but it was true, %v", err)
	}
}

func TestLt(t *testing.T) {
	warrant := Warrant{
		ObjectType:  "role",
		ObjectId:    "admin",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-a",
		Policy:      Policy("amount < 50"),
		CreatedAt:   time.Now(),
	}

	match, err := warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  49,
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if match || err != nil {
		t.Fatalf("Expected match to be false, but it was true, %v", err)
	}
}

func TestLte(t *testing.T) {
	warrant := Warrant{
		ObjectType:  "role",
		ObjectId:    "admin",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-a",
		Policy:      Policy("amount <= 50"),
		CreatedAt:   time.Now(),
	}

	match, err := warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  49,
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  51,
	})
	if match || err != nil {
		t.Fatalf("Expected match to be false, but it was true, %v", err)
	}
}

func TestMatches(t *testing.T) {
	warrant := Warrant{
		ObjectType:  "role",
		ObjectId:    "admin",
		Relation:    "member",
		SubjectType: "user",
		SubjectId:   "user-a",
		Policy:      Policy("firstName matches \"jane\""),
		CreatedAt:   time.Now(),
	}

	match, err := warrant.Policy.Eval(PolicyContext{
		"warrant":   &warrant,
		"firstName": "jane",
		"lastName":  "doe",
	})
	if !match || err != nil {
		t.Fatalf("Expected match to be true, but it was false, %v", err)
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant":   &warrant,
		"firstName": "john",
		"lastName":  "doe",
	})
	if match || err != nil {
		t.Fatalf("Expected match to be false, but it was true, %v", err)
	}
}
