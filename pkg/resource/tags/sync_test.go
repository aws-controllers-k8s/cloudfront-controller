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

package tags

import (
	"context"
	"reflect"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	svcapitypes "github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
)

type fakeTagsClient struct {
	tagged   []svcsdktypes.Tag
	untagged []string
	calls    int
}

func (f *fakeTagsClient) TagResource(
	_ context.Context,
	in *svcsdk.TagResourceInput,
	_ ...func(*svcsdk.Options),
) (*svcsdk.TagResourceOutput, error) {
	f.calls++
	f.tagged = append(f.tagged, in.Tags.Items...)
	return &svcsdk.TagResourceOutput{}, nil
}

func (f *fakeTagsClient) ListTagsForResource(
	_ context.Context,
	_ *svcsdk.ListTagsForResourceInput,
	_ ...func(*svcsdk.Options),
) (*svcsdk.ListTagsForResourceOutput, error) {
	return &svcsdk.ListTagsForResourceOutput{
		Tags: &svcsdktypes.Tags{Items: []svcsdktypes.Tag{}},
	}, nil
}

func (f *fakeTagsClient) UntagResource(
	_ context.Context,
	in *svcsdk.UntagResourceInput,
	_ ...func(*svcsdk.Options),
) (*svcsdk.UntagResourceOutput, error) {
	f.calls++
	f.untagged = append(f.untagged, in.TagKeys.Items...)
	return &svcsdk.UntagResourceOutput{}, nil
}

type fakeMetricsRecorder struct{}

func (fakeMetricsRecorder) RecordAPICall(string, string, error) {}

func strPtr(s string) *string {
	return &s
}

func tag(k, v string) *svcapitypes.Tag {
	return &svcapitypes.Tag{Key: strPtr(k), Value: strPtr(v)}
}

func TestSyncResourceTags(t *testing.T) {
	tests := []struct {
		name         string
		latest       []*svcapitypes.Tag
		desired      []*svcapitypes.Tag
		wantTagged   []svcsdktypes.Tag
		wantUntagged []string
		wantCalls    int
	}{
		{
			name:    "no-op when tags are identical",
			latest:  []*svcapitypes.Tag{tag("a", "1"), tag("b", "2")},
			desired: []*svcapitypes.Tag{tag("b", "2"), tag("a", "1")},
		},
		{
			name:       "add new tags",
			latest:     []*svcapitypes.Tag{tag("a", "1")},
			desired:    []*svcapitypes.Tag{tag("a", "1"), tag("b", "2")},
			wantTagged: []svcsdktypes.Tag{{Key: strPtr("b"), Value: strPtr("2")}},
			wantCalls:  1,
		},
		{
			name:         "remove tags",
			latest:       []*svcapitypes.Tag{tag("a", "1"), tag("b", "2")},
			desired:      []*svcapitypes.Tag{tag("a", "1")},
			wantUntagged: []string{"b"},
			wantCalls:    1,
		},
		// GetTagsDifference treats a tag whose value changed as an addition
		// (upsert semantics): it is re-sent via TagResource and is NOT untagged
		// first. This preserves the behavior of the previous computeTagsDelta
		// implementation, which also reported value changes only in
		// addedOrUpdated and never in removed.
		{
			name:       "value change is tagged without untagging",
			latest:     []*svcapitypes.Tag{tag("a", "1")},
			desired:    []*svcapitypes.Tag{tag("a", "2")},
			wantTagged: []svcsdktypes.Tag{{Key: strPtr("a"), Value: strPtr("2")}},
			wantCalls:  1,
		},
		{
			name:         "add, remove and update together",
			latest:       []*svcapitypes.Tag{tag("a", "1"), tag("b", "2")},
			desired:      []*svcapitypes.Tag{tag("a", "9"), tag("c", "3")},
			wantTagged:   []svcsdktypes.Tag{{Key: strPtr("a"), Value: strPtr("9")}, {Key: strPtr("c"), Value: strPtr("3")}},
			wantUntagged: []string{"b"},
			wantCalls:    2,
		},
		{
			name:    "nil tags do not panic",
			latest:  []*svcapitypes.Tag{nil, {Key: nil, Value: strPtr("x")}, tag("a", "1")},
			desired: []*svcapitypes.Tag{tag("a", "1"), nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeTagsClient{}
			err := SyncResourceTags(
				context.Background(),
				client,
				fakeMetricsRecorder{},
				"arn:aws:cloudfront::123456789012:distribution/EDFDVBD6EXAMPLE",
				tt.latest,
				tt.desired,
			)
			if err != nil {
				t.Fatalf("SyncResourceTags() returned error: %v", err)
			}
			if client.calls != tt.wantCalls {
				t.Errorf("API calls = %d, want %d", client.calls, tt.wantCalls)
			}
			if !reflect.DeepEqual(client.tagged, tt.wantTagged) {
				t.Errorf("tagged = %v, want %v", client.tagged, tt.wantTagged)
			}
			if !reflect.DeepEqual(client.untagged, tt.wantUntagged) {
				t.Errorf("untagged = %v, want %v", client.untagged, tt.wantUntagged)
			}
		})
	}
}

func TestGetResourceTagsEmpty(t *testing.T) {
	got, err := GetResourceTags(
		context.Background(),
		&fakeTagsClient{},
		fakeMetricsRecorder{},
		"arn:aws:cloudfront::123456789012:distribution/EDFDVBD6EXAMPLE",
	)
	if err != nil {
		t.Fatalf("GetResourceTags() returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetResourceTags() = %v, want empty", got)
	}
}
