// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package persistence

import (
	"encoding/json"

	"github.com/gogo/protobuf/types"
	commonproto "go.temporal.io/temporal-proto/common"

	"github.com/temporalio/temporal/common/persistence/serialization"

	"github.com/temporalio/temporal/common"
	"github.com/temporalio/temporal/common/log"
	"github.com/temporalio/temporal/common/log/tag"
)

type (
	visibilityManagerImpl struct {
		serializer  PayloadSerializer
		persistence VisibilityStore
		logger      log.Logger
	}
)

// VisibilityEncoding is default encoding for visibility data
const VisibilityEncoding = common.EncodingTypeThriftRW

var _ VisibilityManager = (*visibilityManagerImpl)(nil)

// NewVisibilityManagerImpl returns new VisibilityManager
func NewVisibilityManagerImpl(persistence VisibilityStore, logger log.Logger) VisibilityManager {
	return &visibilityManagerImpl{
		serializer:  NewPayloadSerializer(),
		persistence: persistence,
		logger:      logger,
	}
}

func (v *visibilityManagerImpl) Close() {
	v.persistence.Close()
}

func (v *visibilityManagerImpl) GetName() string {
	return v.persistence.GetName()
}

func (v *visibilityManagerImpl) RecordWorkflowExecutionStarted(request *RecordWorkflowExecutionStartedRequest) error {
	req := &InternalRecordWorkflowExecutionStartedRequest{
		DomainUUID:         request.DomainUUID,
		WorkflowID:         request.Execution.GetWorkflowId(),
		RunID:              request.Execution.GetRunId(),
		WorkflowTypeName:   request.WorkflowTypeName,
		StartTimestamp:     request.StartTimestamp,
		ExecutionTimestamp: request.ExecutionTimestamp,
		WorkflowTimeout:    request.WorkflowTimeout,
		TaskID:             request.TaskID,
		Memo:               v.serializeMemo(request.Memo, request.DomainUUID, request.Execution.GetWorkflowId(), request.Execution.GetRunId()),
		SearchAttributes:   request.SearchAttributes,
	}
	return v.persistence.RecordWorkflowExecutionStarted(req)
}

func (v *visibilityManagerImpl) RecordWorkflowExecutionClosed(request *RecordWorkflowExecutionClosedRequest) error {
	req := &InternalRecordWorkflowExecutionClosedRequest{
		DomainUUID:         request.DomainUUID,
		WorkflowID:         request.Execution.GetWorkflowId(),
		RunID:              request.Execution.GetRunId(),
		WorkflowTypeName:   request.WorkflowTypeName,
		StartTimestamp:     request.StartTimestamp,
		ExecutionTimestamp: request.ExecutionTimestamp,
		TaskID:             request.TaskID,
		Memo:               v.serializeMemo(request.Memo, request.DomainUUID, request.Execution.GetWorkflowId(), request.Execution.GetRunId()),
		SearchAttributes:   request.SearchAttributes,
		CloseTimestamp:     request.CloseTimestamp,
		Status:             request.Status,
		HistoryLength:      request.HistoryLength,
		RetentionSeconds:   request.RetentionSeconds,
	}
	return v.persistence.RecordWorkflowExecutionClosed(req)
}

func (v *visibilityManagerImpl) UpsertWorkflowExecution(request *UpsertWorkflowExecutionRequest) error {
	req := &InternalUpsertWorkflowExecutionRequest{
		DomainUUID:         request.DomainUUID,
		WorkflowID:         request.Execution.GetWorkflowId(),
		RunID:              request.Execution.GetRunId(),
		WorkflowTypeName:   request.WorkflowTypeName,
		StartTimestamp:     request.StartTimestamp,
		ExecutionTimestamp: request.ExecutionTimestamp,
		TaskID:             request.TaskID,
		Memo:               v.serializeMemo(request.Memo, request.DomainUUID, request.Execution.GetWorkflowId(), request.Execution.GetRunId()),
		SearchAttributes:   request.SearchAttributes,
	}
	return v.persistence.UpsertWorkflowExecution(req)
}

func (v *visibilityManagerImpl) ListOpenWorkflowExecutions(request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListOpenWorkflowExecutions(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutions(request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutions(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListOpenWorkflowExecutionsByType(request *ListWorkflowExecutionsByTypeRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListOpenWorkflowExecutionsByType(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutionsByType(request *ListWorkflowExecutionsByTypeRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutionsByType(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListOpenWorkflowExecutionsByWorkflowID(request *ListWorkflowExecutionsByWorkflowIDRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListOpenWorkflowExecutionsByWorkflowID(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutionsByWorkflowID(request *ListWorkflowExecutionsByWorkflowIDRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutionsByWorkflowID(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutionsByStatus(request *ListClosedWorkflowExecutionsByStatusRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutionsByStatus(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) GetClosedWorkflowExecution(request *GetClosedWorkflowExecutionRequest) (*GetClosedWorkflowExecutionResponse, error) {
	internalResp, err := v.persistence.GetClosedWorkflowExecution(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalGetResponse(internalResp), nil
}

func (v *visibilityManagerImpl) DeleteWorkflowExecution(request *VisibilityDeleteWorkflowExecutionRequest) error {
	return v.persistence.DeleteWorkflowExecution(request)
}

func (v *visibilityManagerImpl) ListWorkflowExecutions(request *ListWorkflowExecutionsRequestV2) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListWorkflowExecutions(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ScanWorkflowExecutions(request *ListWorkflowExecutionsRequestV2) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ScanWorkflowExecutions(request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) CountWorkflowExecutions(request *CountWorkflowExecutionsRequest) (*CountWorkflowExecutionsResponse, error) {
	return v.persistence.CountWorkflowExecutions(request)
}

func (v *visibilityManagerImpl) convertInternalGetResponse(internalResp *InternalGetClosedWorkflowExecutionResponse) *GetClosedWorkflowExecutionResponse {
	if internalResp == nil {
		return nil
	}

	resp := &GetClosedWorkflowExecutionResponse{}
	resp.Execution = v.convertVisibilityWorkflowExecutionInfo(internalResp.Execution)
	return resp
}

func (v *visibilityManagerImpl) convertInternalListResponse(internalResp *InternalListWorkflowExecutionsResponse) *ListWorkflowExecutionsResponse {
	if internalResp == nil {
		return nil
	}

	resp := &ListWorkflowExecutionsResponse{}
	resp.Executions = make([]*commonproto.WorkflowExecutionInfo, len(internalResp.Executions))
	for i, execution := range internalResp.Executions {
		resp.Executions[i] = v.convertVisibilityWorkflowExecutionInfo(execution)
	}

	resp.NextPageToken = internalResp.NextPageToken
	return resp
}

func (v *visibilityManagerImpl) getSearchAttributes(attr map[string]interface{}) (*commonproto.SearchAttributes, error) {
	indexedFields := make(map[string][]byte)
	var err error
	var valBytes []byte
	for k, val := range attr {
		valBytes, err = json.Marshal(val)
		if err != nil {
			v.logger.Error("error when encode search attributes", tag.Value(val))
			continue
		}
		indexedFields[k] = valBytes
	}
	if err != nil {
		return nil, err
	}
	return &commonproto.SearchAttributes{
		IndexedFields: indexedFields,
	}, nil
}

func (v *visibilityManagerImpl) convertVisibilityWorkflowExecutionInfo(execution *VisibilityWorkflowExecutionInfo) *commonproto.WorkflowExecutionInfo {
	// special handling of ExecutionTime for cron or retry
	if execution.ExecutionTime.UnixNano() == 0 {
		execution.ExecutionTime = execution.StartTime
	}

	memo, err := v.serializer.DeserializeVisibilityMemo(execution.Memo)
	if err != nil {
		v.logger.Error("failed to deserialize memo",
			tag.WorkflowID(execution.WorkflowID),
			tag.WorkflowRunID(execution.RunID),
			tag.Error(err))
	}
	searchAttributes, err := v.getSearchAttributes(execution.SearchAttributes)
	if err != nil {
		v.logger.Error("failed to convert search attributes",
			tag.WorkflowID(execution.WorkflowID),
			tag.WorkflowRunID(execution.RunID),
			tag.Error(err))
	}

	convertedExecution := &commonproto.WorkflowExecutionInfo{
		Execution: &commonproto.WorkflowExecution{
			WorkflowId: execution.WorkflowID,
			RunId:      execution.RunID,
		},
		Type: &commonproto.WorkflowType{
			Name: execution.TypeName,
		},
		StartTime: &types.Int64Value{
			Value: execution.StartTime.UnixNano(),
		},
		ExecutionTime:    execution.ExecutionTime.UnixNano(),
		Memo:             memo,
		SearchAttributes: searchAttributes,
	}

	// for close records
	if execution.Status != nil {
		convertedExecution.CloseTime = &types.Int64Value{
			Value: execution.CloseTime.UnixNano(),
		}
		convertedExecution.CloseStatus = *execution.Status
		convertedExecution.HistoryLength = execution.HistoryLength
	}

	return convertedExecution
}

func (v *visibilityManagerImpl) serializeMemo(visibilityMemo *commonproto.Memo, domainID, wID, rID string) *serialization.DataBlob {
	memo, err := v.serializer.SerializeVisibilityMemo(visibilityMemo, VisibilityEncoding)
	if err != nil {
		v.logger.WithTags(
			tag.WorkflowDomainID(domainID),
			tag.WorkflowID(wID),
			tag.WorkflowRunID(rID),
			tag.Error(err)).
			Error("Unable to encode visibility memo")
	}
	if memo == nil {
		return &serialization.DataBlob{}
	}
	return memo
}
