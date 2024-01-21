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

package origin_request_policy

import (
	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
)

// setQuantityFields simply goes through the input shape and sets the Quantity
// field for all list container parent shapes to the length of the Items field.
// This is necessary because CloudFront's API will return an
// `InconsistentQuantities` error message if Quantity != len(Items). This is
// why we can't have nice things, apparently.
func setQuantityFields(dc *svcsdk.OriginRequestPolicyConfig) {
	if dc == nil {
		return
	}
	if dc.CookiesConfig != nil && dc.CookiesConfig.Cookies != nil {
		dc.CookiesConfig.Cookies.SetQuantity(int64(len(dc.CookiesConfig.Cookies.Items)))
	}
	if dc.HeadersConfig != nil && dc.HeadersConfig.Headers != nil {
		dc.HeadersConfig.Headers.SetQuantity(int64(len(dc.HeadersConfig.Headers.Items)))
	}
	if dc.QueryStringsConfig != nil && dc.QueryStringsConfig.QueryStrings != nil {
		dc.QueryStringsConfig.QueryStrings.SetQuantity(int64(len(dc.QueryStringsConfig.QueryStrings.Items)))
	}
}
