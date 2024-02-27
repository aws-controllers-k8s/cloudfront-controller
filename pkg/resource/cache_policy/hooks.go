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

package cache_policy

import (
	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
)

// setQuantityFields simply goes through the input shape and sets the Quantity
// field for all list container parent shapes to the length of the Items field.
// This is necessary because CloudFront's API will return an
// `InconsistentQuantities` error message if Quantity != len(Items). This is
// why we can't have nice things, apparently.
func setQuantityFields(cp *svcsdk.CachePolicyConfig) {
	if cp == nil {
		return
	}
	if cp.ParametersInCacheKeyAndForwardedToOrigin != nil {
		cpp := cp.ParametersInCacheKeyAndForwardedToOrigin
		if cpp.CookiesConfig != nil {
			if cpp.CookiesConfig.Cookies != nil {
				cpp.CookiesConfig.Cookies.SetQuantity(int64(len(cpp.CookiesConfig.Cookies.Items)))
			}
		}
		if cpp.HeadersConfig != nil {
			if cpp.HeadersConfig.Headers != nil {
				cpp.HeadersConfig.Headers.SetQuantity(int64(len(cpp.HeadersConfig.Headers.Items)))
			}
		}
		if cpp.QueryStringsConfig != nil {
			if cpp.QueryStringsConfig.QueryStrings != nil {
				cpp.QueryStringsConfig.QueryStrings.SetQuantity(int64(len(cpp.QueryStringsConfig.QueryStrings.Items)))
			}
		}
	}
}
