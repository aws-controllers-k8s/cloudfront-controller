// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package function

import (
	"context"

	"github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// setResourceAdditionalFields sets any additional fields that are not returned
// by the Describe API operation.
func (rm *resourceManager) setResourceAdditionalFields(ctx context.Context, r *v1alpha1.Function) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceAdditionalFields")
	defer exit(err)

	err = rm.setFunctionCode(ctx, r)
	if err != nil {
		return err
	}
	return nil
}

// setFunctionCode retrieves the function code from the CloudFront API and
// stores it in the FunctionCode field of the supplied Function.
func (rm *resourceManager) setFunctionCode(ctx context.Context, r *v1alpha1.Function) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setFunctionCode")
	defer exit(err)

	output, err := rm.sdkapi.GetFunction(
		ctx,
		&svcsdk.GetFunctionInput{
			Name:  r.Spec.Name,
			Stage: svcsdktypes.FunctionStage(*r.Status.FunctionSummary.FunctionMetadata.Stage),
		},
	)
	rm.metrics.RecordAPICall("GET", "GetFunction", err)
	if err != nil {
		return err
	}
	r.Spec.FunctionCode = []byte(output.FunctionCode)
	return nil
}
