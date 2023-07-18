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
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

// TODO: remove doneC from arg? do we need to check both doneC in arg and task.doneC?
func execute(id int, ctx context.Context, doneC <-chan struct{}, tasksC <-chan Task) {
	for task := range tasksC {
		select {
		case _, ok := <-doneC:
			if !ok {
				log.Ctx(ctx).Debug().Msgf("worker %d exited", id)
				return
			}
			log.Ctx(ctx).Debug().Msgf("worker %d received some value from doneC", id)
		case _, ok := <-task.Done():
			if !ok {
				log.Ctx(ctx).Debug().Msgf("worker %d skipped [%s]", id, task)
				continue
			}
			log.Ctx(ctx).Debug().Msgf("worker %d did not skip", id)
		default:
			start := time.Now()
			task.Execute()
			log.Ctx(ctx).Debug().Msgf("worker %d executed [%s] [%s]", id, task, time.Since(start))
		}
	}
}

// TODO: confirm this doesn't add tasks if doneC is closed (maybe replace with context check)
func addTasks(doneC <-chan struct{}, tasksC chan<- Task, tasks ...Task) {
	go func() {
		select {
		case _, ok := <-doneC:
			if !ok {
				log.Debug().Msgf("addTasks skipped adding %d tasks", len(tasks))
				return
			}
			log.Debug().Msgf("addTasks did not skip adding tasks")
		default:
			for _, t := range tasks {
				tasksC <- t
				//log.Debug().Msgf("addTasks added [%s]", t)
			}
		}
	}()
}

// TODO: should childResultC be len of totalTasks? Probably
// TODO: should we close child channels?
func anyOfChan(resultC chan<- Result, totalTasks int) (chan Result, chan struct{}) {
	childResultC := make(chan Result)
	childDoneC := make(chan struct{})
	//log.Debug().Msgf("START anyOf chan")
	go func() {
		defer close(childDoneC)
		resultsReceived := 0
		for result := range childResultC {
			if result.Err != nil {
				//log.Debug().Err(result.Err).Msgf("ERR anyOf")
				resultC <- result
				return
			}
			if result.Matched {
				//log.Debug().Msgf("MATCH anyOf")
				resultC <- result
				return
			}
			resultsReceived++
			if resultsReceived == totalTasks {
				//log.Debug().Msgf("NO MATCH anyOf")
				resultC <- Result{
					Matched:      false,
					DecisionPath: result.DecisionPath,
					Err:          nil,
				}
				return
			}
		}
	}()
	return childResultC, childDoneC
}

func allOfChan(resultC chan<- Result, totalTasks int) (chan Result, chan struct{}) {
	childResultC := make(chan Result)
	childDoneC := make(chan struct{})
	go func() {
		defer close(childDoneC)
		resultsReceived := 0
		for result := range childResultC {
			if result.Err != nil {
				log.Debug().Err(result.Err).Msgf("ERR allOf")
				resultC <- result
				return
			}
			if !result.Matched {
				resultC <- result
				return
			}
			resultsReceived++
			if resultsReceived == totalTasks {
				resultC <- Result{
					Matched:      true,
					DecisionPath: result.DecisionPath,
					Err:          nil,
				}
				return
			}
		}
	}()
	return childResultC, childDoneC
}

func noneOfChan(resultC chan<- Result, totalTasks int) (chan Result, chan struct{}) {
	childResultC := make(chan Result)
	childDoneC := make(chan struct{})
	go func() {
		defer close(childDoneC)
		resultsReceived := 0
		for result := range childResultC {
			if result.Err != nil {
				log.Debug().Err(result.Err).Msgf("ERR noneOf")
				resultC <- result
				return
			}
			if result.Matched {
				resultC <- Result{
					Matched:      false,
					DecisionPath: result.DecisionPath,
					Err:          nil,
				}
				return
			}
			resultsReceived++
			if resultsReceived == totalTasks {
				resultC <- Result{
					Matched:      true,
					DecisionPath: result.DecisionPath,
					Err:          nil,
				}
				return
			}
		}
	}()
	return childResultC, childDoneC
}

const (
	ResultMatched    = "ResultMatched"
	ResultNotMatched = "ResultNotMatched"
	ResultErr        = "ResultError"
)

type Result struct {
	Matched      bool
	DecisionPath []warrant.WarrantSpec
	Err          error
}

// TODO: replace Done() with ctx?
type Task interface {
	Execute()
	Done() <-chan struct{}
}

type CheckTask struct {
	Level       int
	Ctx         context.Context
	CheckSpec   *CheckSpec
	CurrentPath []warrant.WarrantSpec
	ResultC     chan<- Result
	TaskC       chan<- Task
	DoneC       <-chan struct{}
	Svc         *CheckService
	TypesMap    *ObjectTypeMap
}

func (c CheckTask) Done() <-chan struct{} {
	return c.DoneC
}

func (c CheckTask) String() string {
	return fmt.Sprintf("CheckTask -> level:%d, checkSpec:%s", c.Level, c.CheckSpec)
}

// TODO: cancel context if a path has been exhausted (or mark it as done)
// TODO: fix decisionPath
// TODO: put a semaphore (n concurrent calls) on top of db calls
func (c CheckTask) Execute() {

	// 1. Check for direct warrant match
	matchedWarrant, err := c.Svc.getWithPolicyMatch(c.Ctx, c.CheckSpec.CheckWarrantSpec)
	if err != nil {
		log.Ctx(c.Ctx).Err(err).Msgf("ERR CheckTask getWithPolicyMatch")
		c.ResultC <- Result{
			Matched:      false,
			DecisionPath: c.CurrentPath,
			Err:          err,
		}
		return
	}
	if matchedWarrant != nil {
		c.ResultC <- Result{
			Matched:      true,
			DecisionPath: append(c.CurrentPath, *matchedWarrant),
			Err:          nil,
		}
		return
	}

	objectTypeSpec := c.TypesMap.GetByTypeId(c.CheckSpec.ObjectType)
	if objectTypeSpec == nil {
		err := fmt.Errorf("objecttype %s not found", c.CheckSpec.ObjectType)
		log.Ctx(c.Ctx).Err(err).Msgf("ERR CheckTask GetByTypeId")
		c.ResultC <- Result{
			Matched:      false,
			DecisionPath: c.CurrentPath,
			Err:          err,
		}
		return
	}

	// 2. Check through indirect/group warrants
	// 3. And/or defined rules for target relation
	if relationRule, ok := objectTypeSpec.Relations[c.CheckSpec.Relation]; ok {
		anyOfResultsC, anyOfDoneC := anyOfChan(c.ResultC, 2)
		addTasks(anyOfDoneC, c.TaskC, GroupCheckTask{
			Level:       c.Level + 1,
			Ctx:         c.Ctx,
			CheckSpec:   c.CheckSpec,
			CurrentPath: c.CurrentPath,
			ResultC:     anyOfResultsC,
			TaskC:       c.TaskC,
			DoneC:       anyOfDoneC,
			Svc:         c.Svc,
			TypesMap:    c.TypesMap,
		}, CheckRuleTask{
			Level:       c.Level + 1,
			Ctx:         c.Ctx,
			CheckSpec:   c.CheckSpec,
			CurrentPath: c.CurrentPath,
			ResultC:     anyOfResultsC,
			TaskC:       c.TaskC,
			DoneC:       anyOfDoneC,
			Svc:         c.Svc,
			TypesMap:    c.TypesMap,
			Rule:        &relationRule,
		})
	} else {
		addTasks(c.DoneC, c.TaskC, GroupCheckTask{
			Level:       c.Level + 1,
			Ctx:         c.Ctx,
			CheckSpec:   c.CheckSpec,
			CurrentPath: c.CurrentPath,
			ResultC:     c.ResultC,
			TaskC:       c.TaskC,
			DoneC:       c.DoneC,
			Svc:         c.Svc,
			TypesMap:    c.TypesMap,
		})
	}
}

type GroupCheckTask struct {
	Level       int
	Ctx         context.Context
	CheckSpec   *CheckSpec
	CurrentPath []warrant.WarrantSpec
	ResultC     chan<- Result
	TaskC       chan<- Task
	DoneC       <-chan struct{}
	Svc         *CheckService
	TypesMap    *ObjectTypeMap
}

func (c GroupCheckTask) Done() <-chan struct{} {
	return c.DoneC
}

func (c GroupCheckTask) String() string {
	return fmt.Sprintf("GroupCheckTask -> level:%d, checkSpec:%s", c.Level, c.CheckSpec)
}

func (c GroupCheckTask) Execute() {
	warrants, err := c.Svc.getMatchingSubjects(c.Ctx, c.TypesMap, c.CheckSpec.ObjectType, c.CheckSpec.ObjectId, c.CheckSpec.Relation, c.CheckSpec.Context)
	if err != nil {
		log.Ctx(c.Ctx).Err(err).Msgf("ERR GroupCheckTask getMatchingSubjects")
		c.ResultC <- Result{
			Matched:      false,
			DecisionPath: c.CurrentPath,
			Err:          err,
		}
		return
	}

	var matchingWarrants []warrant.WarrantSpec
	for _, w := range warrants {
		if w.Subject.Relation == "" {
			continue
		}
		matchingWarrants = append(matchingWarrants, w)
	}
	if len(matchingWarrants) == 0 {
		c.ResultC <- Result{
			Matched:      false,
			DecisionPath: c.CurrentPath,
			Err:          nil,
		}
	} else if len(matchingWarrants) == 1 {
		matchingWarrant := matchingWarrants[0]
		addTasks(c.DoneC, c.TaskC, CheckTask{
			Level: c.Level + 1,
			Ctx:   c.Ctx,
			CheckSpec: &CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: matchingWarrant.Subject.ObjectType,
					ObjectId:   matchingWarrant.Subject.ObjectId,
					Relation:   matchingWarrant.Subject.Relation,
					Subject:    c.CheckSpec.Subject,
					Context:    c.CheckSpec.Context,
				},
				Debug: c.CheckSpec.Debug,
			},
			CurrentPath: append(c.CurrentPath, matchingWarrant),
			ResultC:     c.ResultC,
			TaskC:       c.TaskC,
			DoneC:       c.DoneC,
			Svc:         c.Svc,
			TypesMap:    c.TypesMap,
		})
	} else {
		anyOfResultsC, anyOfDoneC := anyOfChan(c.ResultC, len(matchingWarrants))
		var additionalTasks []Task
		for _, matchingWarrant := range matchingWarrants {
			additionalTasks = append(additionalTasks, CheckTask{
				Level: c.Level + 1,
				Ctx:   c.Ctx,
				CheckSpec: &CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: matchingWarrant.Subject.ObjectType,
						ObjectId:   matchingWarrant.Subject.ObjectId,
						Relation:   matchingWarrant.Subject.Relation,
						Subject:    c.CheckSpec.Subject,
						Context:    c.CheckSpec.Context,
					},
					Debug: c.CheckSpec.Debug,
				},
				CurrentPath: append(c.CurrentPath, matchingWarrant),
				ResultC:     anyOfResultsC,
				TaskC:       c.TaskC,
				DoneC:       anyOfDoneC,
				Svc:         c.Svc,
				TypesMap:    c.TypesMap,
			})
		}
		addTasks(c.DoneC, c.TaskC, additionalTasks...)
	}
}

type CheckRuleTask struct {
	Level       int
	Ctx         context.Context
	CheckSpec   *CheckSpec
	CurrentPath []warrant.WarrantSpec
	ResultC     chan<- Result
	TaskC       chan<- Task
	DoneC       <-chan struct{}
	Svc         *CheckService
	TypesMap    *ObjectTypeMap
	Rule        *objecttype.RelationRule
}

func (t CheckRuleTask) Done() <-chan struct{} {
	return t.DoneC
}

func (t CheckRuleTask) String() string {
	return fmt.Sprintf("CheckRuleTask -> level:%d, checkSpec:%s, rule:%s", t.Level, t.CheckSpec, t.Rule)
}

func (t CheckRuleTask) Execute() {
	warrantSpec := t.CheckSpec.CheckWarrantSpec
	if t.Rule == nil {
		t.ResultC <- Result{
			Matched:      false,
			DecisionPath: t.CurrentPath,
			Err:          nil,
		}
		return
	}
	switch t.Rule.InheritIf {
	case "":
		// No match found
		t.ResultC <- Result{
			Matched:      false,
			DecisionPath: t.CurrentPath,
			Err:          nil,
		}
	case objecttype.InheritIfAllOf:
		allOfResultC, allOfDoneC := allOfChan(t.ResultC, len(t.Rule.Rules))
		var allOfTasks []Task
		for _, r := range t.Rule.Rules {
			rule := r
			allOfTasks = append(allOfTasks, CheckRuleTask{
				Level:       t.Level + 1,
				Ctx:         t.Ctx,
				CheckSpec:   t.CheckSpec,
				CurrentPath: t.CurrentPath,
				ResultC:     allOfResultC,
				TaskC:       t.TaskC,
				DoneC:       allOfDoneC,
				Svc:         t.Svc,
				TypesMap:    t.TypesMap,
				Rule:        &rule,
			})
		}
		addTasks(t.DoneC, t.TaskC, allOfTasks...)
	case objecttype.InheritIfAnyOf:
		anyOfResultC, anyOfDoneC := anyOfChan(t.ResultC, len(t.Rule.Rules))
		var anyOfTasks []Task
		for _, r := range t.Rule.Rules {
			rule := r
			anyOfTasks = append(anyOfTasks, CheckRuleTask{
				Level:       t.Level + 1,
				Ctx:         t.Ctx,
				CheckSpec:   t.CheckSpec,
				CurrentPath: t.CurrentPath,
				ResultC:     anyOfResultC,
				TaskC:       t.TaskC,
				DoneC:       anyOfDoneC,
				Svc:         t.Svc,
				TypesMap:    t.TypesMap,
				Rule:        &rule,
			})
		}
		addTasks(t.DoneC, t.TaskC, anyOfTasks...)
	case objecttype.InheritIfNoneOf:
		noneOfResultC, noneOfDoneC := noneOfChan(t.ResultC, len(t.Rule.Rules))
		var noneOfTasks []Task
		for _, r := range t.Rule.Rules {
			rule := r
			noneOfTasks = append(noneOfTasks, CheckRuleTask{
				Level:       t.Level + 1,
				Ctx:         t.Ctx,
				CheckSpec:   t.CheckSpec,
				CurrentPath: t.CurrentPath,
				ResultC:     noneOfResultC,
				TaskC:       t.TaskC,
				DoneC:       noneOfDoneC,
				Svc:         t.Svc,
				TypesMap:    t.TypesMap,
				Rule:        &rule,
			})
		}
		addTasks(t.DoneC, t.TaskC, noneOfTasks...)
	default:
		if t.Rule.OfType == "" && t.Rule.WithRelation == "" {
			addTasks(t.DoneC, t.TaskC, CheckTask{
				Level: t.Level + 1,
				Ctx:   t.Ctx,
				CheckSpec: &CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: warrantSpec.ObjectType,
						ObjectId:   warrantSpec.ObjectId,
						Relation:   t.Rule.InheritIf,
						Subject:    warrantSpec.Subject,
						Context:    warrantSpec.Context,
					},
					Debug: t.CheckSpec.Debug,
				},
				CurrentPath: t.CurrentPath,
				ResultC:     t.ResultC,
				TaskC:       t.TaskC,
				DoneC:       t.DoneC,
				Svc:         t.Svc,
				TypesMap:    t.TypesMap,
			})
			return
		}

		matchingWarrants, err := t.Svc.getMatchingSubjectsBySubjectType(t.Ctx, t.TypesMap, warrantSpec.ObjectType, warrantSpec.ObjectId, t.Rule.WithRelation, t.Rule.OfType, warrantSpec.Context)
		if err != nil {
			log.Ctx(t.Ctx).Err(err).Msgf("ERR CheckRuleTask getMatchingSubjectsBySubjectType")
			t.ResultC <- Result{
				Matched:      false,
				DecisionPath: t.CurrentPath,
				Err:          err,
			}
			return
		}

		if len(matchingWarrants) == 0 {
			t.ResultC <- Result{
				Matched:      false,
				DecisionPath: t.CurrentPath,
				Err:          nil,
			}
		} else if len(matchingWarrants) == 1 {
			matchingWarrant := matchingWarrants[0]
			addTasks(t.DoneC, t.TaskC, CheckTask{
				Level: t.Level + 1,
				Ctx:   t.Ctx,
				CheckSpec: &CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: matchingWarrant.Subject.ObjectType,
						ObjectId:   matchingWarrant.Subject.ObjectId,
						Relation:   t.Rule.InheritIf,
						Subject:    warrantSpec.Subject,
						Context:    warrantSpec.Context,
					},
					Debug: t.CheckSpec.Debug,
				},
				CurrentPath: append(t.CurrentPath, matchingWarrant),
				ResultC:     t.ResultC,
				TaskC:       t.TaskC,
				DoneC:       t.DoneC,
				Svc:         t.Svc,
				TypesMap:    t.TypesMap,
			})
		} else {
			subResultsC, subDoneC := anyOfChan(t.ResultC, len(matchingWarrants))
			var additionalTasks []Task
			for _, matchingWarrant := range matchingWarrants {
				additionalTasks = append(additionalTasks, CheckTask{
					Level: t.Level + 1,
					Ctx:   t.Ctx,
					CheckSpec: &CheckSpec{
						CheckWarrantSpec: CheckWarrantSpec{
							ObjectType: matchingWarrant.Subject.ObjectType,
							ObjectId:   matchingWarrant.Subject.ObjectId,
							Relation:   t.Rule.InheritIf,
							Subject:    warrantSpec.Subject,
							Context:    warrantSpec.Context,
						},
						Debug: t.CheckSpec.Debug,
					},
					CurrentPath: append(t.CurrentPath, matchingWarrant),
					ResultC:     subResultsC,
					TaskC:       t.TaskC,
					DoneC:       subDoneC,
					Svc:         t.Svc,
					TypesMap:    t.TypesMap,
				})
			}
			addTasks(t.DoneC, t.TaskC, additionalTasks...)
		}
	}
}
