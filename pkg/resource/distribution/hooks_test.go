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

package distribution

import (
	"fmt"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"k8s.io/apimachinery/pkg/api/equality"

	svcapitypes "github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
)

func testDistribution(mutate func(*svcapitypes.DistributionConfig)) *resource {
	functionAssociations := func() *svcapitypes.FunctionAssociations {
		return &svcapitypes.FunctionAssociations{Items: []*svcapitypes.FunctionAssociation{
			{EventType: aws.String("viewer-request"), FunctionARN: aws.String("arn:aws:cloudfront::111122223333:function/request")},
			{EventType: aws.String("viewer-response"), FunctionARN: aws.String("arn:aws:cloudfront::111122223333:function/response")},
		}}
	}
	lambdaFunctionAssociations := func() *svcapitypes.LambdaFunctionAssociations {
		return &svcapitypes.LambdaFunctionAssociations{Items: []*svcapitypes.LambdaFunctionAssociation{
			{EventType: aws.String("origin-request"), LambdaFunctionARN: aws.String("arn:aws:lambda:us-east-1:111122223333:function:request:1")},
			{EventType: aws.String("origin-response"), LambdaFunctionARN: aws.String("arn:aws:lambda:us-east-1:111122223333:function:response:1")},
		}}
	}
	members := func(ids ...string) *svcapitypes.OriginGroupMembers {
		items := []*svcapitypes.OriginGroupMember{}
		for _, id := range ids {
			items = append(items, &svcapitypes.OriginGroupMember{OriginID: aws.String(id)})
		}
		return &svcapitypes.OriginGroupMembers{Items: items}
	}
	dc := &svcapitypes.DistributionConfig{
		DefaultCacheBehavior: &svcapitypes.DefaultCacheBehavior{
			TargetOriginID:             aws.String("s3"),
			FunctionAssociations:       functionAssociations(),
			LambdaFunctionAssociations: lambdaFunctionAssociations(),
		},
		CacheBehaviors: &svcapitypes.CacheBehaviors{Items: []*svcapitypes.CacheBehavior{
			{
				PathPattern:    aws.String("/index.html"),
				TargetOriginID: aws.String("s3"),
				AllowedMethods: &svcapitypes.AllowedMethods{
					Items:         aws.StringSlice([]string{"GET", "HEAD", "OPTIONS"}),
					CachedMethods: &svcapitypes.CachedMethods{Items: aws.StringSlice([]string{"GET", "HEAD"})},
				},
				FunctionAssociations:       functionAssociations(),
				LambdaFunctionAssociations: lambdaFunctionAssociations(),
			},
			{
				PathPattern:    aws.String("/api/*"),
				TargetOriginID: aws.String("api"),
			},
		}},
		Origins: &svcapitypes.Origins{Items: []*svcapitypes.Origin{
			{ID: aws.String("s3"), DomainName: aws.String("bucket.s3.us-east-1.amazonaws.com")},
			{ID: aws.String("api"), DomainName: aws.String("api.example.com")},
			{ID: aws.String("backup"), DomainName: aws.String("backup.example.com")},
		}},
		OriginGroups: &svcapitypes.OriginGroups{Items: []*svcapitypes.OriginGroup{
			{ID: aws.String("api-group"), Members: members("api", "backup")},
			{ID: aws.String("s3-group"), Members: members("s3", "backup")},
		}},
	}
	if mutate != nil {
		mutate(dc)
	}
	return &resource{ko: &svcapitypes.Distribution{Spec: svcapitypes.DistributionSpec{DistributionConfig: dc}}}
}

func TestCustomPreCompare(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*svcapitypes.DistributionConfig)
		// wantDiffAt is the path expected to differ, or "" for no difference.
		wantDiffAt string
	}{
		{
			name: "identical",
		},
		{
			name: "function associations in another order",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				slices.Reverse(dc.DefaultCacheBehavior.FunctionAssociations.Items)
				slices.Reverse(dc.DefaultCacheBehavior.LambdaFunctionAssociations.Items)
				slices.Reverse(dc.CacheBehaviors.Items[0].FunctionAssociations.Items)
				slices.Reverse(dc.CacheBehaviors.Items[0].LambdaFunctionAssociations.Items)
			},
		},
		{
			name: "allowed and cached methods in another order",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors.Items[0].AllowedMethods.Items = aws.StringSlice([]string{"HEAD", "GET", "OPTIONS"})
				dc.CacheBehaviors.Items[0].AllowedMethods.CachedMethods.Items = aws.StringSlice([]string{"HEAD", "GET"})
			},
		},
		{
			name: "origins and origin groups in another order",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				slices.Reverse(dc.Origins.Items)
				slices.Reverse(dc.OriginGroups.Items)
			},
		},
		{
			name: "default behavior function association changed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.DefaultCacheBehavior.FunctionAssociations.Items[1].FunctionARN = aws.String("arn:aws:cloudfront::111122223333:function/other")
			},
			wantDiffAt: "Spec.DistributionConfig.DefaultCacheBehavior.FunctionAssociations.Items",
		},
		{
			name: "default behavior Lambda function association removed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.DefaultCacheBehavior.LambdaFunctionAssociations.Items = dc.DefaultCacheBehavior.LambdaFunctionAssociations.Items[:1]
			},
			wantDiffAt: "Spec.DistributionConfig.DefaultCacheBehavior.LambdaFunctionAssociations.Items",
		},
		{
			name: "cache behavior function association changed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors.Items[0].FunctionAssociations.Items[0].EventType = aws.String("viewer-response")
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors.Items",
		},
		{
			name: "cache behavior Lambda function association changed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors.Items[0].LambdaFunctionAssociations.Items[0].IncludeBody = aws.Bool(true)
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors.Items",
		},
		{
			name: "allowed method removed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors.Items[0].AllowedMethods.Items = aws.StringSlice([]string{"HEAD", "GET"})
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors.Items",
		},
		{
			name: "cached method added",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors.Items[0].AllowedMethods.CachedMethods.Items = aws.StringSlice([]string{"GET", "HEAD", "OPTIONS"})
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors.Items",
		},
		{
			name: "cache behaviors in another order",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				slices.Reverse(dc.CacheBehaviors.Items)
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors.Items",
		},
		{
			name: "origin changed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.Origins.Items[1].DomainName = aws.String("api-2.example.com")
			},
			wantDiffAt: "Spec.DistributionConfig.Origins.Items",
		},
		{
			name: "origin removed",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.Origins.Items = dc.Origins.Items[:2]
			},
			wantDiffAt: "Spec.DistributionConfig.Origins.Items",
		},
		{
			name: "origin group members in another order",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				slices.Reverse(dc.OriginGroups.Items[0].Members.Items)
			},
			wantDiffAt: "Spec.DistributionConfig.OriginGroups.Items",
		},
		// A list missing on one side is left to the generated nil checks.
		{
			name: "cache behaviors missing",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors = nil
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors",
		},
		{
			name: "default behavior function associations missing",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.DefaultCacheBehavior.FunctionAssociations = nil
			},
			wantDiffAt: "Spec.DistributionConfig.DefaultCacheBehavior.FunctionAssociations",
		},
		{
			name: "origins missing",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.Origins = nil
			},
			wantDiffAt: "Spec.DistributionConfig.Origins",
		},
		{
			name: "origin groups missing",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.OriginGroups = nil
			},
			wantDiffAt: "Spec.DistributionConfig.OriginGroups",
		},
		{
			name: "nil cache behavior",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.CacheBehaviors.Items[1] = nil
			},
			wantDiffAt: "Spec.DistributionConfig.CacheBehaviors.Items",
		},
		{
			name: "nil function association",
			mutate: func(dc *svcapitypes.DistributionConfig) {
				dc.DefaultCacheBehavior.FunctionAssociations.Items[0] = nil
			},
			wantDiffAt: "Spec.DistributionConfig.DefaultCacheBehavior.FunctionAssociations.Items",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired := testDistribution(nil)
			latest := testDistribution(tt.mutate)
			delta := newResourceDelta(desired, latest)
			paths := []string{}
			for _, difference := range delta.Differences {
				paths = append(paths, fmt.Sprint(difference.Path))
			}
			if tt.wantDiffAt == "" && delta.DifferentAt("Spec") {
				t.Fatalf("expected no difference, got differences at %v", paths)
			}
			if tt.wantDiffAt != "" && !delta.DifferentAt(tt.wantDiffAt) {
				t.Fatalf("expected a difference at %s, got differences at %v", tt.wantDiffAt, paths)
			}
			// The comparison must not reorder either resource.
			if !equality.Semantic.DeepEqual(desired.ko.Spec, testDistribution(nil).ko.Spec) ||
				!equality.Semantic.DeepEqual(latest.ko.Spec, testDistribution(tt.mutate).ko.Spec) {
				t.Fatal("comparison modified a compared resource")
			}
		})
	}
}
