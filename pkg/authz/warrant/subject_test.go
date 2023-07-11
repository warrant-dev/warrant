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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToMapDirectSubjectSpec(t *testing.T) {
	subject := SubjectSpec{
		ObjectType: "user",
		ObjectId:   "user-A",
	}
	expectedMap := map[string]interface{}{
		"objectType": "user",
		"objectId":   "user-A",
	}
	actualMap := subject.ToMap()
	if !cmp.Equal(actualMap, expectedMap) {
		t.Fatalf("Expected subject string to be %s, but it was %s", expectedMap, actualMap)
	}
}

func TestToMapGroupSubjectSpec(t *testing.T) {
	subject := SubjectSpec{
		ObjectType: "role",
		ObjectId:   "admin",
		Relation:   "member",
	}
	expectedMap := map[string]interface{}{
		"objectType": "role",
		"objectId":   "admin",
		"relation":   "member",
	}
	actualMap := subject.ToMap()
	if !cmp.Equal(actualMap, expectedMap) {
		t.Fatalf("Expected subject string to be %s, but it was %s", expectedMap, actualMap)
	}
}

func TestToStringDirectSubjectSpec(t *testing.T) {
	subject := SubjectSpec{
		ObjectType: "user",
		ObjectId:   "user-A",
	}
	expectedSubjectStr := "user:user-A"
	actualSubjectStr := subject.String()
	if actualSubjectStr != expectedSubjectStr {
		t.Fatalf("Expected subject string to be %s, but it was %s", expectedSubjectStr, actualSubjectStr)
	}
}

func TestToStringGroupSubjectSpec(t *testing.T) {
	subject := SubjectSpec{
		ObjectType: "role",
		ObjectId:   "admin",
		Relation:   "member",
	}
	expectedSubjectStr := "role:admin#member"
	actualSubjectStr := subject.String()
	if actualSubjectStr != expectedSubjectStr {
		t.Fatalf("Expected subject string to be %s, but it was %s", expectedSubjectStr, actualSubjectStr)
	}
}

func TestStringToSubjectSpecDirectSubjectSpec(t *testing.T) {
	subjectStr := "user:user-A"
	expectedSubjectSpec := &SubjectSpec{
		ObjectType: "user",
		ObjectId:   "user-A",
	}
	actualSubjectSpec, err := StringToSubjectSpec(subjectStr)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualSubjectSpec, expectedSubjectSpec) {
		t.Fatalf("Expected subject spec to be %v, but it was %v", expectedSubjectSpec, actualSubjectSpec)
	}
}

func TestStringToSubjectSpecGroupSubjectSpec(t *testing.T) {
	subjectStr := "role:admin#member"
	expectedSubjectSpec := &SubjectSpec{
		ObjectType: "role",
		ObjectId:   "admin",
		Relation:   "member",
	}
	actualSubjectSpec, err := StringToSubjectSpec(subjectStr)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(actualSubjectSpec, expectedSubjectSpec) {
		t.Fatalf("Expected subject spec to be %v, but it was %v", expectedSubjectSpec, actualSubjectSpec)
	}
}

func TestStringToSubjectSpecMultiplePounds(t *testing.T) {
	subjectStr := "role:admin#member#"
	expectedErrStr := fmt.Sprintf("invalid subject string %s", subjectStr)
	_, err := StringToSubjectSpec(subjectStr)
	if err == nil || err.Error() != expectedErrStr {
		t.Fatalf("Expected err to be %s, but it was %v", expectedErrStr, err)
	}
}

func TestStringToSubjectSpecNoColon(t *testing.T) {
	subjectStr := "roleadmin#member"
	expectedErrStr := fmt.Sprintf("invalid subject string %s", subjectStr)
	_, err := StringToSubjectSpec(subjectStr)
	if err == nil || err.Error() != expectedErrStr {
		t.Fatalf("Expected err to be %s, but it was %v", expectedErrStr, err)
	}
}
