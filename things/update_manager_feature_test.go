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

package things

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/update-manager/api/types"
	"github.com/eclipse-kanto/update-manager/test"
	"github.com/eclipse-kanto/update-manager/test/mocks"
	"github.com/eclipse/ditto-clients-golang/model"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/eclipse/ditto-clients-golang/protocol/things"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const outboxPathFmt = "/features/UpdateManager/outbox/messages/%s"

var (
	tesThingID = model.NewNamespacedIDFrom("namespace:testDevice")
	testWG     = &sync.WaitGroup{}
	errTest    = fmt.Errorf("test error")
)

func TestActivate(t *testing.T) {
	tests := map[string]struct {
		feature       *updateManagerFeature
		mockExecution func(*mocks.MockClient) error
	}{
		"test_activate_ok": {
			feature: &updateManagerFeature{thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Subscribe(gomock.Any())
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).DoAndReturn(func(message *protocol.Envelope) error {
					assert.False(t, message.Headers.IsResponseRequired())
					assertTwinCommandTopic(t, *tesThingID, message.Topic)
					assert.Equal(t, "/features/UpdateManager", message.Path)
					feature := message.Value.(*model.Feature)
					assert.Equal(t, updateManagerFeatureDefinition, feature.Definition[0].String())
					assert.Equal(t, test.Domain, feature.Properties["domain"])
					return nil
				})
				return nil
			},
		},
		"test_activate_already_activated": {
			feature: &updateManagerFeature{active: true},
			mockExecution: func(_ *mocks.MockClient) error {
				return nil
			},
		},
		"test_activate_error": {
			feature: &updateManagerFeature{thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Subscribe(gomock.Any())
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).Return(errTest)
				mockDittoClient.EXPECT().Unsubscribe()
				return errTest
			},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl, mockDittoClient, _, _ := setupMocks(t, testCase.feature)
			defer mockCtrl.Finish()

			expectedError := testCase.mockExecution(mockDittoClient)
			actualError := testCase.feature.Activate()
			if expectedError != nil {
				assert.EqualError(t, actualError, expectedError.Error())
				assert.False(t, testCase.feature.active)
			} else {
				assert.Nil(t, actualError)
				assert.True(t, testCase.feature.active)
			}
		})
	}
}

func TestDeactivate(t *testing.T) {
	tests := map[string]struct {
		feature       *updateManagerFeature
		mockExecution func(*mocks.MockClient)
	}{
		"test_deactivate_ok": {
			feature: &updateManagerFeature{active: true},
			mockExecution: func(mockDittoClient *mocks.MockClient) {
				mockDittoClient.EXPECT().Unsubscribe()
			},
		},
		"test_deactivate_already_deactivated": {
			feature:       &updateManagerFeature{},
			mockExecution: func(_ *mocks.MockClient) {},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl, mockDittoClient, _, _ := setupMocks(t, testCase.feature)
			defer mockCtrl.Finish()

			testCase.mockExecution(mockDittoClient)
			testCase.feature.Deactivate()

			assert.False(t, testCase.feature.active)
		})
	}
}

func TestSetState(t *testing.T) {
	tests := map[string]struct {
		feature       *updateManagerFeature
		mockExecution func(*mocks.MockClient) error
	}{
		"test_set_state_ok": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).DoAndReturn(func(message *protocol.Envelope) error {
					assert.False(t, message.Headers.IsResponseRequired())
					assertTwinCommandTopic(t, *tesThingID, message.Topic)
					assert.Equal(t, "/features/UpdateManager/properties", message.Path)
					properties := message.Value.(*updateManagerProperties)
					assert.Equal(t, test.Domain, properties.Domain)
					assert.Equal(t, test.ActivityID, properties.ActivityID)
					assert.Equal(t, test.Inventory, properties.Inventory)
					assert.True(t, properties.Timestamp > 0)
					return nil
				})
				return nil
			},
		},
		"test_set_state_not_active": {
			feature: &updateManagerFeature{active: false},
			mockExecution: func(_ *mocks.MockClient) error {
				return nil
			},
		},
		"test_set_state_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).Return(errTest)
				return errTest
			},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl, mockDittoClient, _, _ := setupMocks(t, testCase.feature)
			defer mockCtrl.Finish()

			expectedError := testCase.mockExecution(mockDittoClient)
			actualError := testCase.feature.SetState(test.ActivityID, test.Inventory)
			if expectedError != nil {
				assert.EqualError(t, actualError, expectedError.Error())
			} else {
				assert.Nil(t, actualError)
			}
		})
	}
}

func TestSendFeedback(t *testing.T) {
	testFeedback := &types.DesiredStateFeedback{}
	tests := map[string]struct {
		feature       *updateManagerFeature
		mockExecution func(*mocks.MockClient) error
	}{
		"test_send_feedback_ok": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).DoAndReturn(func(message *protocol.Envelope) error {
					assert.False(t, message.Headers.IsResponseRequired())
					assertLiveMessageTopic(t, *tesThingID, updateManagerFeatureMessageFeedback, message.Topic)
					assert.Equal(t, fmt.Sprintf(outboxPathFmt, updateManagerFeatureMessageFeedback), message.Path)
					feedback := message.Value.(*feedback)
					assert.Equal(t, test.ActivityID, feedback.ActivityID)
					assert.Equal(t, testFeedback, feedback.DesiredStateFeedback)
					assert.True(t, feedback.Timestamp > 0)
					return nil
				})
				return nil
			},
		},
		"test_send_feedback_not_active": {
			feature: &updateManagerFeature{active: false},
			mockExecution: func(_ *mocks.MockClient) error {
				return nil
			},
		},
		"test_send_feedback_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).Return(errTest)
				return errTest
			},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl, mockDittoClient, _, _ := setupMocks(t, testCase.feature)
			defer mockCtrl.Finish()

			expectedError := testCase.mockExecution(mockDittoClient)
			actualError := testCase.feature.SendFeedback(test.ActivityID, testFeedback)
			if expectedError != nil {
				assert.EqualError(t, actualError, expectedError.Error())
			} else {
				assert.Nil(t, actualError)
			}
		})
	}
}

func TestSendConsent(t *testing.T) {
	testConsent := &types.OwnerConsent{}
	tests := map[string]struct {
		feature       *updateManagerFeature
		mockExecution func(*mocks.MockClient) error
	}{
		"test_send_consent_ok": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).DoAndReturn(func(message *protocol.Envelope) error {
					assert.True(t, message.Headers.IsResponseRequired())
					assertLiveMessageTopic(t, *tesThingID, updateManagerFeatureMessageConsent, message.Topic)
					assert.Equal(t, fmt.Sprintf(outboxPathFmt, updateManagerFeatureMessageConsent), message.Path)
					consent := message.Value.(*consent)
					assert.Equal(t, test.ActivityID, consent.ActivityID)
					assert.Equal(t, testConsent, consent.OwnerConsent)
					assert.True(t, consent.Timestamp > 0)
					return nil
				})
				return nil
			},
		},
		"test_send_consent_not_active": {
			feature: &updateManagerFeature{active: false},
			mockExecution: func(_ *mocks.MockClient) error {
				return nil
			},
		},
		"test_send_consent_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID, domain: test.Domain},
			mockExecution: func(mockDittoClient *mocks.MockClient) error {
				mockDittoClient.EXPECT().Send(gomock.AssignableToTypeOf(&protocol.Envelope{})).Return(errTest)
				return errTest
			},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl, mockDittoClient, _, _ := setupMocks(t, testCase.feature)
			defer mockCtrl.Finish()

			expectedError := testCase.mockExecution(mockDittoClient)
			actualError := testCase.feature.SendConsent(test.ActivityID, testConsent)
			if expectedError != nil {
				assert.EqualError(t, actualError, expectedError.Error())
			} else {
				assert.Nil(t, actualError)
			}
		})
	}
}

func TestSetConsentHandler(t *testing.T) {
	testFeature := &updateManagerFeature{}
	testFeature.SetConsentHandler(mocks.NewMockOwnerConsentHandler(gomock.NewController(t)))
	assert.NotNil(t, testFeature.consentHandler)
	testFeature.SetConsentHandler(nil)
	assert.Nil(t, testFeature.consentHandler)
}

func TestMessageHandler(t *testing.T) {
	testRequestID := "testRequestID"
	testConsentFeedback := &types.OwnerConsentFeedback{
		Status: types.StatusApproved,
	}
	mockThingExecution := func(operation string) func(*mocks.MockClient, *mocks.MockUpdateAgentHandler, *mocks.MockOwnerConsentHandler) {
		return func(mockDittoClient *mocks.MockClient, mockHandler *mocks.MockUpdateAgentHandler, mockConsentHandler *mocks.MockOwnerConsentHandler) {
			mockDittoClient.EXPECT().Reply(testRequestID, gomock.AssignableToTypeOf(&protocol.Envelope{})).DoAndReturn(
				func(_ string, message *protocol.Envelope) error {
					assert.False(t, message.Headers.IsResponseRequired())
					assert.Equal(t, 204, message.Status)
					assertLiveMessageTopic(t, *tesThingID, protocol.TopicAction(operation), message.Topic)
					assert.Equal(t, fmt.Sprintf(outboxPathFmt, operation), message.Path)
					assert.Nil(t, message.Value)
					return nil
				})
			testWG.Add(1)
			switch operation {
			case updateManagerFeatureOperationRefresh:
				mockHandler.EXPECT().HandleCurrentStateGet(test.ActivityID, gomock.Any()).DoAndReturn(func(activityID string, timestamp int64) error {
					testWG.Done()
					return nil
				})
			case updateManagerFeatureOperationApply:
				mockHandler.EXPECT().HandleDesiredState(test.ActivityID, gomock.Any(), gomock.Any()).DoAndReturn(func(activityID string, timestamp int64, ds *types.DesiredState) error {
					assert.Equal(t, test.DesiredState, ds)
					testWG.Done()
					return nil
				})
			case updateManagerFeatureMessageConsent:
				mockConsentHandler.EXPECT().HandleOwnerConsentFeedback(test.ActivityID, gomock.Any(), gomock.Any()).DoAndReturn(func(activityID string, timestamp int64, cf *types.OwnerConsentFeedback) error {
					assert.Equal(t, testConsentFeedback, cf)
					testWG.Done()
					return nil
				})
			default:
				testWG.Done()
			}
		}
	}

	mockThingErrorExecution := func(operation string) func(*mocks.MockClient, *mocks.MockUpdateAgentHandler, *mocks.MockOwnerConsentHandler) {
		return func(mockDittoClient *mocks.MockClient, _ *mocks.MockUpdateAgentHandler, _ *mocks.MockOwnerConsentHandler) {
			mockDittoClient.EXPECT().Reply(testRequestID, gomock.AssignableToTypeOf(&protocol.Envelope{})).DoAndReturn(
				func(_ string, message *protocol.Envelope) error {
					assert.False(t, message.Headers.IsResponseRequired())
					assert.Equal(t, responseStatusBadRequest, message.Status)
					assertLiveMessageTopic(t, *tesThingID, protocol.TopicAction(operation), message.Topic)
					assert.Equal(t, fmt.Sprintf(outboxPathFmt, operation), message.Path)
					thingError := message.Value.(*thingError)
					assert.Equal(t, messagesParameterInvalid, thingError.ErrorCode)
					assert.Equal(t, responseStatusBadRequest, thingError.Status)
					assert.NotEmpty(t, thingError.Message)
					return nil
				})
		}
	}

	tests := map[string]struct {
		feature       *updateManagerFeature
		envelope      *protocol.Envelope
		mockExecution func(*mocks.MockClient, *mocks.MockUpdateAgentHandler, *mocks.MockOwnerConsentHandler)
	}{
		"test_message_handler_not_active": {
			feature:       &updateManagerFeature{},
			mockExecution: func(_ *mocks.MockClient, _ *mocks.MockUpdateAgentHandler, _ *mocks.MockOwnerConsentHandler) {},
		},
		"test_message_handler_unexpected_command": {
			feature:       &updateManagerFeature{active: true, thingID: tesThingID},
			envelope:      things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox("unexpected").Envelope(),
			mockExecution: func(_ *mocks.MockClient, _ *mocks.MockUpdateAgentHandler, _ *mocks.MockOwnerConsentHandler) {},
		},
		"test_message_handler_unexpected_thing_id": {
			feature:       &updateManagerFeature{active: true, thingID: tesThingID},
			envelope:      things.NewMessage(model.NewNamespacedIDFrom("ns:unexpected")).Feature(updateManagerFeatureID).Inbox("unexpected").Envelope(),
			mockExecution: func(_ *mocks.MockClient, _ *mocks.MockUpdateAgentHandler, _ *mocks.MockOwnerConsentHandler) {},
		},
		"test_message_handler_refresh_ok": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureOperationRefresh).WithPayload(&base{ActivityID: test.ActivityID}).
				Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingExecution(updateManagerFeatureOperationRefresh),
		},
		"test_message_handler_refresh_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureOperationRefresh).WithPayload("invalid payload").
				Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingErrorExecution(updateManagerFeatureOperationRefresh),
		},
		"test_message_handler_apply_ok": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureOperationApply).WithPayload(&applyArgs{base: base{ActivityID: test.ActivityID}, DesiredState: test.DesiredState}).
				Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingExecution(updateManagerFeatureOperationApply),
		},
		"test_message_handler_apply_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureOperationApply).WithPayload("invalid payload").
				Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingErrorExecution(updateManagerFeatureOperationApply),
		},
		"test_message_handler_apply_nil_desire_state_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureOperationApply).WithPayload(&applyArgs{base: base{ActivityID: test.ActivityID}}).
				Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingErrorExecution(updateManagerFeatureOperationApply),
		},
		"test_message_handler_consent_ok": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureMessageConsent).
				WithPayload(&consentFeedback{base: base{ActivityID: test.ActivityID}, OwnerConsentFeedback: testConsentFeedback}).Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingExecution(updateManagerFeatureMessageConsent),
		},
		"test_message_handler_consent_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureMessageConsent).WithPayload("invalid payload").
				Envelope(protocol.WithResponseRequired(true)),
			mockExecution: mockThingErrorExecution(updateManagerFeatureMessageConsent),
		},
		"test_message_handler_correlation_id_mismatch_error": {
			feature: &updateManagerFeature{active: true, thingID: tesThingID},
			envelope: things.NewMessage(tesThingID).Feature(updateManagerFeatureID).Inbox(updateManagerFeatureMessageConsent).
				WithPayload(&consentFeedback{base: base{ActivityID: test.ActivityID}, OwnerConsentFeedback: testConsentFeedback}).Envelope(protocol.WithResponseRequired(true), protocol.WithCorrelationID("mismatch")),
			mockExecution: mockThingErrorExecution(updateManagerFeatureMessageConsent),
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl, mockDittoClient, mockHandler, mockConsentHandler := setupMocks(t, testCase.feature)
			defer mockCtrl.Finish()

			testCase.mockExecution(mockDittoClient, mockHandler, mockConsentHandler)
			testCase.feature.messagesHandler(testRequestID, testCase.envelope)
			test.AssertWithTimeout(t, testWG, 2*time.Second)
		})
	}
}

func assertTwinCommandTopic(t *testing.T, tesThingID model.NamespacedID, topic *protocol.Topic) {
	expectedTopic := (&protocol.Topic{}).
		WithNamespace(tesThingID.Namespace).
		WithEntityName(tesThingID.Name).
		WithGroup(protocol.GroupThings).
		WithChannel(protocol.ChannelTwin).
		WithCriterion(protocol.CriterionCommands).
		WithAction(protocol.ActionModify)

	assert.Equal(t, expectedTopic, topic)
}

func assertLiveMessageTopic(t *testing.T, tesThingID model.NamespacedID, operation protocol.TopicAction, topic *protocol.Topic) {
	expectedTopic := (&protocol.Topic{}).
		WithNamespace(tesThingID.Namespace).
		WithEntityName(tesThingID.Name).
		WithGroup(protocol.GroupThings).
		WithChannel(protocol.ChannelLive).
		WithCriterion(protocol.CriterionMessages).
		WithAction(operation)

	assert.Equal(t, expectedTopic, topic)
}

func setupMocks(t *testing.T, feature *updateManagerFeature) (*gomock.Controller, *mocks.MockClient, *mocks.MockUpdateAgentHandler, *mocks.MockOwnerConsentHandler) {
	mockCtrl := gomock.NewController(t)
	mockDittoClient := mocks.NewMockClient(mockCtrl)
	mockHandler := mocks.NewMockUpdateAgentHandler(mockCtrl)
	mockConsentHandler := mocks.NewMockOwnerConsentHandler(mockCtrl)
	feature.dittoClient = mockDittoClient
	feature.handler = mockHandler
	feature.consentHandler = mockConsentHandler
	return mockCtrl, mockDittoClient, mockHandler, mockConsentHandler
}
