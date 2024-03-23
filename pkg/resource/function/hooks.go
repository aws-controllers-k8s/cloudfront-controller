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
	"errors"
	"strconv"

	"github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func customPreCompare(delta *ackcompare.Delta, a, b *resource) {
	if a.ko != nil && functionAutoPublishEnabled(a.ko) &&
		b.ko != nil && needsPublish(b.ko) {
		delta.Add("Spec.PublishFunction", nil, nil)
	}
}

// needsPublish returns true if the DEVELOPMENT version of the function has
// not been published to the LIVE stage. It compares the DEVELOPMENT ETag
// (Status.ETag) against the LIVE ETag (Status.LiveETag). If LiveETag is nil,
// the function has never been published.
func needsPublish(f *v1alpha1.Function) bool {
	if f.Status.ETag == nil {
		return false
	}
	if f.Status.LiveETag == nil {
		return true
	}
	return *f.Status.ETag != *f.Status.LiveETag
}

// setResourceAdditionalFields sets any additional fields that are not returned
// by the Describe API operation.
func (rm *resourceManager) setResourceAdditionalFields(ctx context.Context, r *v1alpha1.Function) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceAdditionalFields")
	defer exit(err)

	if err = rm.setFunctionCode(ctx, r); err != nil {
		return err
	}
	if err = rm.setLiveETag(ctx, r); err != nil {
		return err
	}
	return nil
}

// setLiveETag describes the LIVE stage of the function and stores its ETag
// in Status.LiveETag. If the function has never been published, LiveETag
// is set to nil.
func (rm *resourceManager) setLiveETag(ctx context.Context, r *v1alpha1.Function) error {
	resp, err := rm.sdkapi.DescribeFunction(
		ctx,
		&svcsdk.DescribeFunctionInput{
			Name:  r.Spec.Name,
			Stage: svcsdktypes.FunctionStageLive,
		},
	)
	rm.metrics.RecordAPICall("GET", "DescribeFunction", err)
	if err != nil {
		var notFound *svcsdktypes.NoSuchFunctionExists
		if errors.As(err, &notFound) {
			r.Status.LiveETag = nil
			return nil
		}
		return err
	}
	r.Status.LiveETag = resp.ETag
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

// publishes a CloudFront function
func (rm *resourceManager) publishFunction(ctx context.Context, r *v1alpha1.Function) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.publishFunction")
	defer func() { exit(err) }()

	_, err = rm.sdkapi.PublishFunction(
		ctx,
		&svcsdk.PublishFunctionInput{
			Name:    r.Spec.Name,
			IfMatch: r.Status.ETag,
		},
	)
	rm.metrics.RecordAPICall("POST", "PublishFunction", err)
	if err != nil {
		return err
	}
	return nil
}

// functionAutoPublishEnabled returns true if the function should be
// automatically published after a successful update or create operation.
func functionAutoPublishEnabled(f *v1alpha1.Function) bool {
	annotations := f.ObjectMeta.GetAnnotations()
	if annotations == nil {
		return false
	}

	autoPublish, ok := annotations[v1alpha1.AutoPublishAnnotation]
	if ok {
		return autoPublish == strconv.FormatBool(true)
	}

	// By default we do not auto-publish functions.
	return false
}
