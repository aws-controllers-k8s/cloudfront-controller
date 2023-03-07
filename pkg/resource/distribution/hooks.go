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
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
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
func setQuantityFields(input *svcsdk.CreateDistributionInput) {
	if input.DistributionConfig == nil {
		return
	}
	dc := input.DistributionConfig
	if dc.Aliases != nil {
		dc.Aliases.SetQuantity(int64(len(dc.Aliases.Items)))
	}
	cbs := dc.CacheBehaviors
	if cbs != nil && cbs.Items != nil {
		for _, cb := range cbs.Items {
			ams := cb.AllowedMethods
			if ams != nil {
				if ams.Items != nil {
					ams.SetQuantity(int64(len(ams.Items)))
				}
				cms := ams.CachedMethods
				if cms != nil && cms.Items != nil {
					cms.SetQuantity(int64(len(cms.Items)))
				}
			}
			fvs := cb.ForwardedValues
			if fvs != nil {
				cks := fvs.Cookies
				if cks != nil && cks.WhitelistedNames != nil {
					wlns := cks.WhitelistedNames
					if wlns.Items != nil {
						wlns.SetQuantity(int64(len(wlns.Items)))
					}
				}
				hds := fvs.Headers
				if hds != nil && hds.Items != nil {
					hds.SetQuantity(int64(len(hds.Items)))
				}
				qscks := fvs.QueryStringCacheKeys
				if qscks != nil && qscks.Items != nil {
					qscks.SetQuantity(int64(len(qscks.Items)))
				}
			}
			fas := cb.FunctionAssociations
			if fas != nil && fas.Items != nil {
				fas.SetQuantity(int64(len(fas.Items)))
			}
			lfas := cb.LambdaFunctionAssociations
			if lfas != nil && lfas.Items != nil {
				lfas.SetQuantity(int64(len(lfas.Items)))
			}
			tkgs := cb.TrustedKeyGroups
			if tkgs != nil && tkgs.Items != nil {
				tkgs.SetQuantity(int64(len(tkgs.Items)))
			}
			tss := cb.TrustedSigners
			if tss != nil && tss.Items != nil {
				tss.SetQuantity(int64(len(tss.Items)))
			}
		}
		cbs.SetQuantity(int64(len(cbs.Items)))
	}
	cers := dc.CustomErrorResponses
	if cers != nil && cers.Items != nil {
		cers.SetQuantity(int64(len(cers.Items)))
	}
	dcb := dc.DefaultCacheBehavior
	if dcb != nil {
		ams := dcb.AllowedMethods
		if ams != nil {
			if ams.Items != nil {
				ams.SetQuantity(int64(len(ams.Items)))
			}
			cms := ams.CachedMethods
			if cms != nil && cms.Items != nil {
				cms.SetQuantity(int64(len(cms.Items)))
			}
		}
		fvs := dcb.ForwardedValues
		if fvs != nil {
			cks := fvs.Cookies
			if cks != nil && cks.WhitelistedNames != nil {
				wlns := cks.WhitelistedNames
				if wlns.Items != nil {
					wlns.SetQuantity(int64(len(wlns.Items)))
				}
			}
			hds := fvs.Headers
			if hds != nil && hds.Items != nil {
				hds.SetQuantity(int64(len(hds.Items)))
			}
			qscks := fvs.QueryStringCacheKeys
			if qscks != nil && qscks.Items != nil {
				qscks.SetQuantity(int64(len(qscks.Items)))
			}
		}
		fas := dcb.FunctionAssociations
		if fas != nil && fas.Items != nil {
			fas.SetQuantity(int64(len(fas.Items)))
		}
		lfas := dcb.LambdaFunctionAssociations
		if lfas != nil && lfas.Items != nil {
			lfas.SetQuantity(int64(len(lfas.Items)))
		}
		tkgs := dcb.TrustedKeyGroups
		if tkgs != nil && tkgs.Items != nil {
			tkgs.SetQuantity(int64(len(tkgs.Items)))
		}
		tss := dcb.TrustedSigners
		if tss != nil && tss.Items != nil {
			tss.SetQuantity(int64(len(tss.Items)))
		}
	}
	ogs := dc.OriginGroups
	if ogs != nil && ogs.Items != nil {
		for _, og := range ogs.Items {
			fc := og.FailoverCriteria
			if fc != nil && fc.StatusCodes != nil {
				scs := fc.StatusCodes
				if scs.Items != nil {
					scs.SetQuantity(int64(len(scs.Items)))
				}
			}
			if og.Members != nil && og.Members.Items != nil {
				og.Members.SetQuantity(int64(len(og.Members.Items)))
			}
		}
		ogs.SetQuantity(int64(len(ogs.Items)))
	}
	os := dc.Origins
	if os != nil && os.Items != nil {
		for _, o := range os.Items {
			chs := o.CustomHeaders
			if chs != nil && chs.Items != nil {
				chs.SetQuantity(int64(len(chs.Items)))
			}
			coc := o.CustomOriginConfig
			if coc != nil {
				osps := coc.OriginSslProtocols
				if osps != nil && osps.Items != nil {
					osps.SetQuantity(int64(len(osps.Items)))
				}
			}
		}
		os.SetQuantity(int64(len(os.Items)))
	}
	rs := dc.Restrictions
	if rs != nil {
		grs := rs.GeoRestriction
		if grs != nil && grs.Items != nil {
			grs.SetQuantity(int64(len(grs.Items)))
		}
	}
}
