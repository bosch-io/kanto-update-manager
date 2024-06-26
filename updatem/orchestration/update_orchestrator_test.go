// Copyright (c) 2023 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

package orchestration

import (
	"context"
	"testing"
	"time"

	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/api/types"
	"github.com/eclipse-kanto/update-manager/config"
	"github.com/eclipse-kanto/update-manager/test"
	"github.com/eclipse-kanto/update-manager/test/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateOrchestrator(t *testing.T) {
	expectedOrchestrator := &updateOrchestrator{
		cfg: &config.Config{
			RebootEnabled: true,
		},
		phaseTimeout:        10 * time.Minute,
		ownerConsentTimeout: 30 * time.Minute,
	}
	assert.Equal(t, expectedOrchestrator, NewUpdateOrchestrator(&config.Config{RebootEnabled: true}, nil))
}

func TestUpdOrchApply(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	doneChan := make(chan bool, 1)
	applyChan := make(chan bool, 1)

	t.Run("test_valid_scenario", func(t *testing.T) {
		eventCallback := mocks.NewMockUpdateManagerCallback(mockCtrl)
		ctx := context.Background()

		updOrchestrator := &updateOrchestrator{
			cfg:          createTestConfig(false, false),
			phaseTimeout: 10 * time.Minute,
		}
		domainAgent := mocks.NewMockUpdateManager(mockCtrl)

		domainAgent.EXPECT().Apply(ctx, test.ActivityID, test.DesiredState).DoAndReturn(func(ctx context.Context, activityId string, desiredState *types.DesiredState) {
			applyChan <- true
		})
		domainAgent.EXPECT().Name().AnyTimes()

		domainAgents := map[string]api.UpdateManager{
			"domain1": domainAgent,
		}

		eventCallback.EXPECT().HandleDesiredStateFeedbackEvent("device", test.ActivityID, "", gomock.Any(), "", []*types.Action{}).Times(3)

		go applyDesiredState(ctx, updOrchestrator, doneChan, domainAgents, test.ActivityID, test.DesiredState, eventCallback)

		<-applyChan
		updOrchestrator.HandleDesiredStateFeedbackEvent("domain1", test.ActivityID, "", types.StatusIdentified, "", []*types.Action{})
		updOrchestrator.HandleDesiredStateFeedbackEvent("domain1", test.ActivityID, "", types.StatusCompleted, "", []*types.Action{})
		updOrchestrator.HandleDesiredStateFeedbackEvent("domain1", test.ActivityID, "", types.BaselineStatusCleanupSuccess, "", []*types.Action{})
		<-doneChan
	})
	t.Run("test_empty_domainAgents_err_not_nil", func(t *testing.T) {
		eventCallback := mocks.NewMockUpdateManagerCallback(mockCtrl)
		updOrchestrator := &updateOrchestrator{
			cfg: createTestConfig(false, false),
		}
		desiredState := &types.DesiredState{
			Domains: []*types.Domain{},
		}

		eventCallback.EXPECT().HandleDesiredStateFeedbackEvent("device", "", "", types.StatusIncomplete, "the desired state manifest does not contain any supported domain", []*types.Action{})

		assert.False(t, updOrchestrator.Apply(context.Background(), nil, test.ActivityID, desiredState, eventCallback))
	})
}

func TestHandleOwnerConsent(t *testing.T) {
	updateOrchestrator := &updateOrchestrator{
		operation: &updateOperation{
			activityID:     test.ActivityID,
			ownerConsented: make(chan bool),
		},
	}
	t.Run("test_handle_owner_approved", func(t *testing.T) {
		go updateOrchestrator.HandleOwnerConsentFeedback(test.ActivityID, 0, &types.OwnerConsentFeedback{Status: types.StatusApproved})
		select {
		case consented := <-updateOrchestrator.operation.ownerConsented:
			assert.True(t, consented)
		case <-time.After(1 * time.Second):
			t.Fatal("owner consent not received")
		}
	})
	t.Run("test_handle_owner_denied", func(t *testing.T) {
		go updateOrchestrator.HandleOwnerConsentFeedback(test.ActivityID, 0, &types.OwnerConsentFeedback{Status: types.StatusDenied})
		select {
		case consented := <-updateOrchestrator.operation.ownerConsented:
			assert.False(t, consented)
		case <-time.After(1 * time.Second):
			t.Fatal("owner consent not received")
		}
	})
	t.Run("test_handle_owner_approved_another_activity", func(t *testing.T) {
		go updateOrchestrator.HandleOwnerConsentFeedback("anotherActivity", 0, &types.OwnerConsentFeedback{Status: types.StatusApproved})
		select {
		case <-updateOrchestrator.operation.ownerConsented:
			t.Fatal("unexpected owner consent")
		case <-time.After(1 * time.Second):
			// do nothing
		}
	})
}

func applyDesiredState(ctx context.Context, updOrch *updateOrchestrator, done chan bool, domainAgents map[string]api.UpdateManager, activityID string, desiredState *types.DesiredState, apiDesState api.DesiredStateFeedbackHandler) {
	updOrch.Apply(ctx, domainAgents, activityID, desiredState, apiDesState)
	done <- true
}
