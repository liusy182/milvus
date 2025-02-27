// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datacoord

import (
	"context"
	"errors"
	"time"

	"github.com/milvus-io/milvus/internal/proto/commonpb"
	"github.com/milvus-io/milvus/internal/util/tsoutil"
)

// Response response interface for verification
type Response interface {
	GetStatus() *commonpb.Status
}

// VerifyResponse verify grpc Response 1. check error is nil 2. check response.GetStatus() with status success
func VerifyResponse(response interface{}, err error) error {
	if err != nil {
		return err
	}
	if response == nil {
		return errNilResponse
	}
	switch resp := response.(type) {
	case Response:
		// note that resp will not be nil here, since it's still an interface
		if resp.GetStatus() == nil {
			return errNilStatusResponse
		}
		if resp.GetStatus().GetErrorCode() != commonpb.ErrorCode_Success {
			return errors.New(resp.GetStatus().GetReason())
		}
	case *commonpb.Status:
		if resp == nil {
			return errNilResponse
		}
		if resp.ErrorCode != commonpb.ErrorCode_Success {
			return errors.New(resp.GetReason())
		}
	default:
		return errUnknownResponseType
	}
	return nil
}

// FailResponse sets status to failed with reason
func FailResponse(status *commonpb.Status, reason string) {
	status.ErrorCode = commonpb.ErrorCode_UnexpectedError
	status.Reason = reason
}

func getTimetravelReverseTime(ctx context.Context, allocator allocator) (*timetravel, error) {
	ts, err := allocator.allocTimestamp(ctx)
	if err != nil {
		return nil, err
	}

	pts, _ := tsoutil.ParseTS(ts)
	ttpts := pts.Add(-time.Duration(Params.DataCoordCfg.RetentionDuration) * time.Second)
	tt := tsoutil.ComposeTS(ttpts.UnixNano()/int64(time.Millisecond), 0)
	return &timetravel{tt}, nil
}
