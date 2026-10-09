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
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"k8s.io/apimachinery/pkg/api/equality"

	svcapitypes "github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/cloudfront-controller/pkg/resource/tags"
)

// getIdempotencyToken returns a unique string to be used in certain API calls
// to ensure no replay of the call.
func getIdempotencyToken() string {
	t := time.Now().UTC()
	return t.Format("20060102150405000000")
}

// setQuantityFields simply goes through the input shape and sets the Quantity
// field for all list container parent shapes to the length of the Items field.
// This is necessary because CloudFront's API will return an
// `InconsistentQuantities` error message if Quantity != len(Items). This is
// why we can't have nice things, apparently.
func setQuantityFields(dc *svcsdktypes.DistributionConfig) {
	if dc.Aliases != nil {
		dc.Aliases.Quantity = aws.Int32(int32(len(dc.Aliases.Items)))
	}
	cbs := dc.CacheBehaviors
	if cbs != nil {
		for _, cb := range cbs.Items {
			ams := cb.AllowedMethods
			if ams != nil {
				if ams.Items != nil {
					ams.Quantity = aws.Int32(int32(len(ams.Items)))
				}
				cms := ams.CachedMethods
				if cms != nil {
					cms.Quantity = aws.Int32(int32(len(cms.Items)))
				}
			}
			fvs := cb.ForwardedValues
			if fvs != nil {
				cks := fvs.Cookies
				if cks != nil && cks.WhitelistedNames != nil {
					wlns := cks.WhitelistedNames
					if wlns.Items != nil {
						wlns.Quantity = aws.Int32(int32(len(wlns.Items)))
					}
				}
				hds := fvs.Headers
				if hds != nil {
					hds.Quantity = aws.Int32(int32(len(hds.Items)))
				}
				qscks := fvs.QueryStringCacheKeys
				if qscks != nil {
					qscks.Quantity = aws.Int32(int32(len(qscks.Items)))
				}
			}
			fas := cb.FunctionAssociations
			if fas != nil {
				fas.Quantity = aws.Int32(int32(len(fas.Items)))
			}
			lfas := cb.LambdaFunctionAssociations
			if lfas != nil {
				lfas.Quantity = aws.Int32(int32(len(lfas.Items)))
			}
			tkgs := cb.TrustedKeyGroups
			if tkgs != nil {
				tkgs.Quantity = aws.Int32(int32(len(tkgs.Items)))
			}
			tss := cb.TrustedSigners
			if tss != nil {
				tss.Quantity = aws.Int32(int32(len(tss.Items)))
			}
		}
		cbs.Quantity = aws.Int32(int32(len(cbs.Items)))
	}
	cers := dc.CustomErrorResponses
	if cers != nil {
		cers.Quantity = aws.Int32(int32(len(cers.Items)))
	}
	dcb := dc.DefaultCacheBehavior
	if dcb != nil {
		ams := dcb.AllowedMethods
		if ams != nil {
			if ams.Items != nil {
				ams.Quantity = aws.Int32(int32(len(ams.Items)))
			}
			cms := ams.CachedMethods
			if cms != nil {
				cms.Quantity = aws.Int32(int32(len(cms.Items)))
			}
		}
		fvs := dcb.ForwardedValues
		if fvs != nil {
			cks := fvs.Cookies
			if cks != nil && cks.WhitelistedNames != nil {
				wlns := cks.WhitelistedNames
				if wlns.Items != nil {
					wlns.Quantity = aws.Int32(int32(len(wlns.Items)))
				}
			}
			hds := fvs.Headers
			if hds != nil {
				hds.Quantity = aws.Int32(int32(len(hds.Items)))
			}
			qscks := fvs.QueryStringCacheKeys
			if qscks != nil {
				qscks.Quantity = aws.Int32(int32(len(qscks.Items)))
			}
		}
		fas := dcb.FunctionAssociations
		if fas != nil {
			fas.Quantity = aws.Int32(int32(len(fas.Items)))
		}
		lfas := dcb.LambdaFunctionAssociations
		if lfas != nil {
			lfas.Quantity = aws.Int32(int32(len(lfas.Items)))
		}
		tkgs := dcb.TrustedKeyGroups
		if tkgs != nil {
			tkgs.Quantity = aws.Int32(int32(len(tkgs.Items)))
		}
		tss := dcb.TrustedSigners
		if tss != nil {
			tss.Quantity = aws.Int32(int32(len(tss.Items)))
		}
	}
	ogs := dc.OriginGroups
	if ogs != nil {
		for _, og := range ogs.Items {
			fc := og.FailoverCriteria
			if fc != nil && fc.StatusCodes != nil {
				scs := fc.StatusCodes
				if scs.Items != nil {
					scs.Quantity = aws.Int32(int32(len(scs.Items)))
				}
			}
			if og.Members != nil {
				og.Members.Quantity = aws.Int32(int32(len(og.Members.Items)))
			}
		}
		ogs.Quantity = aws.Int32(int32(len(ogs.Items)))
	}
	os := dc.Origins
	if os != nil {
		for _, o := range os.Items {
			chs := o.CustomHeaders
			if chs != nil {
				chs.Quantity = aws.Int32(int32(len(chs.Items)))
			}
			coc := o.CustomOriginConfig
			if coc != nil {
				osps := coc.OriginSslProtocols
				if osps != nil {
					osps.Quantity = aws.Int32(int32(len(osps.Items)))
				}
			}
		}
		os.Quantity = aws.Int32(int32(len(os.Items)))
	}
	rs := dc.Restrictions
	if rs != nil {
		grs := rs.GeoRestriction
		if grs != nil {
			grs.Quantity = aws.Int32(int32(len(grs.Items)))
		}
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

// distributionDeployed returns true if the supplied distribution is in an active status
func distributionDeployed(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	ds := *r.ko.Status.Status
	return ds == "Deployed"
}

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the distribution cannot be modified until it reaches an deployed
// status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.Status == nil {
		return nil
	}
	status := *r.ko.Status.Status
	return ackrequeue.NeededAfter(
		fmt.Errorf("distribution in '%s' state, cannot be modified until '%s'",
			status, "Deployed"),
		ackrequeue.DefaultRequeueAfterDuration,
	)
}

// customPreCompare compares the Distribution fields that hold lists CloudFront
// does not keep in order. The generated comparison ignores these fields (see
// generator.yaml). CloudFront stores the function associations of a cache
// behavior in an arbitrary order on every write, and returns allowed methods
// and origins in an order of its own. A behavior with two functions would
// match only about half the time, so with several of them nearly every resync
// would call UpdateDistribution. Each field is compared after sorting the
// lists whose order has no meaning.
func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	dcA := a.ko.Spec.DistributionConfig
	dcB := b.ko.Spec.DistributionConfig
	if dcA == nil || dcB == nil {
		return
	}
	if dcA.CacheBehaviors != nil && dcB.CacheBehaviors != nil {
		compareItems(delta, "Spec.DistributionConfig.CacheBehaviors.Items",
			dcA.CacheBehaviors.Items, dcB.CacheBehaviors.Items, normalizeCacheBehaviors)
	}
	if dcbA, dcbB := dcA.DefaultCacheBehavior, dcB.DefaultCacheBehavior; dcbA != nil && dcbB != nil {
		if dcbA.FunctionAssociations != nil && dcbB.FunctionAssociations != nil {
			compareItems(delta, "Spec.DistributionConfig.DefaultCacheBehavior.FunctionAssociations.Items",
				dcbA.FunctionAssociations.Items, dcbB.FunctionAssociations.Items, sortFunctionAssociations)
		}
		if dcbA.LambdaFunctionAssociations != nil && dcbB.LambdaFunctionAssociations != nil {
			compareItems(delta, "Spec.DistributionConfig.DefaultCacheBehavior.LambdaFunctionAssociations.Items",
				dcbA.LambdaFunctionAssociations.Items, dcbB.LambdaFunctionAssociations.Items, sortLambdaFunctionAssociations)
		}
	}
	if dcA.OriginGroups != nil && dcB.OriginGroups != nil {
		compareItems(delta, "Spec.DistributionConfig.OriginGroups.Items",
			dcA.OriginGroups.Items, dcB.OriginGroups.Items, sortOriginGroups)
	}
	if dcA.Origins != nil && dcB.Origins != nil {
		compareItems(delta, "Spec.DistributionConfig.Origins.Items",
			dcA.Origins.Items, dcB.Origins.Items, sortOrigins)
	}
}

// compareItems adds a difference at path when a and b differ once both are
// normalized.
func compareItems[T any](
	delta *ackcompare.Delta,
	path string,
	a []*T,
	b []*T,
	normalize func([]*T) []*T,
) {
	if len(a) != len(b) {
		delta.Add(path, a, b)
	} else if len(a) > 0 && !equality.Semantic.Equalities.DeepEqual(normalize(a), normalize(b)) {
		delta.Add(path, a, b)
	}
}

// normalizeCacheBehaviors returns copies of the cache behaviors with the lists
// inside each behavior sorted. The behaviors themselves keep their order: it
// is their precedence.
func normalizeCacheBehaviors(
	behaviors []*svcapitypes.CacheBehavior,
) []*svcapitypes.CacheBehavior {
	normalized := make([]*svcapitypes.CacheBehavior, len(behaviors))
	for i, behavior := range behaviors {
		if behavior == nil {
			continue
		}
		behavior = behavior.DeepCopy()
		if behavior.AllowedMethods != nil {
			behavior.AllowedMethods.Items = sortMethods(behavior.AllowedMethods.Items)
			if behavior.AllowedMethods.CachedMethods != nil {
				behavior.AllowedMethods.CachedMethods.Items = sortMethods(behavior.AllowedMethods.CachedMethods.Items)
			}
		}
		if behavior.FunctionAssociations != nil {
			behavior.FunctionAssociations.Items = sortFunctionAssociations(behavior.FunctionAssociations.Items)
		}
		if behavior.LambdaFunctionAssociations != nil {
			behavior.LambdaFunctionAssociations.Items = sortLambdaFunctionAssociations(behavior.LambdaFunctionAssociations.Items)
		}
		normalized[i] = behavior
	}
	return normalized
}

// A cache behavior has at most one function association and one Lambda
// function association per event type.
func sortFunctionAssociations(
	items []*svcapitypes.FunctionAssociation,
) []*svcapitypes.FunctionAssociation {
	return sortedBy(items, func(item *svcapitypes.FunctionAssociation) *string { return item.EventType })
}

func sortLambdaFunctionAssociations(
	items []*svcapitypes.LambdaFunctionAssociation,
) []*svcapitypes.LambdaFunctionAssociation {
	return sortedBy(items, func(item *svcapitypes.LambdaFunctionAssociation) *string { return item.EventType })
}

func sortMethods(methods []*string) []*string {
	return sortedBy(methods, func(method *string) *string { return method })
}

// Origins and origin groups are referenced by ID. The members of an origin
// group keep their order: it is their failover order.
func sortOrigins(items []*svcapitypes.Origin) []*svcapitypes.Origin {
	return sortedBy(items, func(item *svcapitypes.Origin) *string { return item.ID })
}

func sortOriginGroups(items []*svcapitypes.OriginGroup) []*svcapitypes.OriginGroup {
	return sortedBy(items, func(item *svcapitypes.OriginGroup) *string { return item.ID })
}

// sortedBy returns a copy of items sorted by key. The elements themselves are
// shared with items, not copied.
func sortedBy[T any](items []*T, key func(*T) *string) []*T {
	value := func(item *T) string {
		if item == nil {
			return ""
		}
		return aws.ToString(key(item))
	}
	sorted := slices.Clone(items)
	slices.SortStableFunc(sorted, func(x, y *T) int {
		return strings.Compare(value(x), value(y))
	})
	return sorted
}
