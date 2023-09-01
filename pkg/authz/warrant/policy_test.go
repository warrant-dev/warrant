// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"strings"
	"testing"
	"time"
)

func TestPolicyContextToString(t *testing.T) {
	t.Parallel()
	policyContext := PolicyContext{
		"hello": "world",
		"user": map[string]interface{}{
			"email": "john.doe@gmail.com",
		},
	}
	expectedPolicyContextStr := "[hello=world user=map[email:john.doe@gmail.com]]"
	actualPolicyContextStr := policyContext.String()
	if actualPolicyContextStr != expectedPolicyContextStr {
		t.Fatalf("Expected policy context string to be %s, but it was %s", expectedPolicyContextStr, actualPolicyContextStr)
	}
}

func TestPolicyEvalUndefinedVariables(t *testing.T) {
	t.Parallel()
	_, err := Policy("a > b").Eval(PolicyContext{})
	if err == nil {
		t.Fatal("Expected err to be non-nil, but it was nil")
	}
}

func TestPolicyEvalSyntaxError(t *testing.T) {
	t.Parallel()
	_, err := Policy("a >").Eval(PolicyContext{})
	if err == nil {
		t.Fatal("Expected err to be non-nil, but it was nil")
	}
}

func TestExpiresIn(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatal("Expected match to be true, but it was false")
	}

	// wait for the policy to expire
	time.Sleep(10 * time.Millisecond)

	match, err = warrant.Policy.Eval(PolicyContext{"warrant": &warrant})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if match {
		t.Fatal("Expected match to be false, but it was true")
	}
}

func TestExpiresInInvalidDuration(t *testing.T) {
	t.Parallel()
	expectedErrStr := "invalid duration string 1"
	_, err := Policy("expiresIn(\"1\")").Eval(PolicyContext{})
	if err == nil || !strings.Contains(err.Error(), expectedErrStr) {
		t.Fatalf("Expected err to be non-nil and contain \"%s\", but it did not: %v", expectedErrStr, err)
	}
}

func TestGt(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}
}

func TestGte(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  49,
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}
}

func TestLt(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}
}

func TestLte(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  50,
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant": &warrant,
		"amount":  51,
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}
}

func TestMatches(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = warrant.Policy.Eval(PolicyContext{
		"warrant":   &warrant,
		"firstName": "john",
		"lastName":  "doe",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}
}

func TestOr(t *testing.T) {
	t.Parallel()
	p := Policy("")

	p = p.Or("firstName == \"jane\"")
	err := p.Validate()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	match, err := p.Eval(PolicyContext{
		"firstName": "jane",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	p = p.Or("lastName == \"doe\"")
	err = p.Validate()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	match, err = p.Eval(PolicyContext{
		"lastName": "doe",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = p.Eval(PolicyContext{
		"firstName": "jane",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	expectedPolicy := "(firstName == \"jane\") || (lastName == \"doe\")"
	if p != Policy(expectedPolicy) {
		t.Fatalf("Expected policy to be %s but got %s", expectedPolicy, p)
	}
}

func TestAnd(t *testing.T) {
	t.Parallel()
	p := Policy("")

	p = p.And("firstName == \"jane\"")
	err := p.Validate()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	match, err := p.Eval(PolicyContext{
		"firstName": "jane",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	p = p.And("lastName == \"doe\"")
	err = p.Validate()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	match, err = p.Eval(PolicyContext{
		"lastName": "doe",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}

	match, err = p.Eval(PolicyContext{
		"firstName": "jane",
		"lastName":  "doe",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	expectedPolicy := "(firstName == \"jane\") && (lastName == \"doe\")"
	if p != Policy(expectedPolicy) {
		t.Fatalf("Expected policy to be %s but got %s", expectedPolicy, p)
	}
}

func TestNot(t *testing.T) {
	t.Parallel()
	p := Policy("")

	newPolicy := Not(p)
	if newPolicy != "" {
		t.Fatalf("Expected empty policy but got %s", newPolicy)
	}

	p = Policy("firstName == \"jane\"")
	p = Not(p)
	match, err := p.Eval(PolicyContext{
		"firstName": "notjane",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !match {
		t.Fatalf("Expected match to be true, but it was false")
	}

	match, err = p.Eval(PolicyContext{
		"firstName": "jane",
	})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if match {
		t.Fatalf("Expected match to be false, but it was true")
	}

	expectedPolicy := "!(firstName == \"jane\")"
	if p != Policy(expectedPolicy) {
		t.Fatalf("Expected policy to be %s but got %s", expectedPolicy, p)
	}
}
