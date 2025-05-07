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

package vpc_origin

import (
	"context"

	svcapitypes "github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/cloudfront-controller/pkg/resource/tags"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// setQuantityFields simply goes through the input shape and sets the Quantity
// field for all list container parent shapes to the length of the Items field.
// This is necessary because CloudFront's API will return an
// `InconsistentQuantities` error message if Quantity != len(Items). This is
// why we can't have nice things, apparently.
func setQuantityFields(ve *svcsdktypes.VpcOriginEndpointConfig) {
	if ve != nil && ve.OriginSslProtocols != nil && ve.OriginSslProtocols.Items != nil {
		ve.OriginSslProtocols.Quantity = aws.Int32(int32(len(ve.OriginSslProtocols.Items)))
	}
}

func updateTagsInCreateRequest(resource *resource, input *svcsdk.CreateVpcOriginInput) {
	input.Tags = nil
	desiredTags := svcsdktypes.Tags{}
	if resource.ko.Spec.Tags != nil {
		requestedTags := []svcsdktypes.Tag{}
		for _, tag := range resource.ko.Spec.Tags {

			tag := svcsdktypes.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			}
			requestedTags = append(requestedTags, tag)
		}

		desiredTags.Items = requestedTags
		input.Tags = &desiredTags
	}
}

// getTags retrieves the resource's associated tags.
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) ([]*svcapitypes.Tag, error) {
	return tags.GetResourceTags(ctx, rm.sdkapi, rm.metrics, resourceARN)
}

// syncTags keeps the resource's tags in sync.
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	return tags.SyncResourceTags(ctx, rm.sdkapi, rm.metrics, string(*latest.ko.Status.ACKResourceMetadata.ARN), desired.ko.Spec.Tags, latest.ko.Spec.Tags)
}
