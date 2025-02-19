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
	"time"

	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	svcapitypes "github.com/aws-controllers-k8s/cloudfront-controller/apis/v1alpha1"
	util "github.com/aws-controllers-k8s/cloudfront-controller/pkg/resource/tags"
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
	return util.GetResourceTags(ctx, rm.sdkapi, rm.metrics, resourceARN)
}

// syncTags keeps the resource's tags in sync.
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	return util.SyncResourceTags(ctx, rm.sdkapi, rm.metrics, string(*latest.ko.Status.ACKResourceMetadata.ARN), desired.ko.Spec.Tags, latest.ko.Spec.Tags)
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
