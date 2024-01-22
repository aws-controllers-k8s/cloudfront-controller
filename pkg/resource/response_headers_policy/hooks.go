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

package response_headers_policy

import (
	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
)

// setQuantityFields simply goes through the input shape and sets the Quantity
// field for all list container parent shapes to the length of the Items field.
// This is necessary because CloudFront's API will return an
// `InconsistentQuantities` error message if Quantity != len(Items). This is
// why we can't have nice things, apparently.
func setQuantityFields(rhp *svcsdk.ResponseHeadersPolicyConfig) {
	if rhp == nil {
		return
	}
	if rhp.CorsConfig != nil {
		if rhp.CorsConfig.AccessControlAllowHeaders != nil && rhp.CorsConfig.AccessControlAllowHeaders.Items != nil {
			rhp.CorsConfig.AccessControlAllowHeaders.SetQuantity(int64(len(rhp.CorsConfig.AccessControlAllowHeaders.Items)))
		}
		if rhp.CorsConfig.AccessControlAllowMethods != nil && rhp.CorsConfig.AccessControlAllowMethods.Items != nil {
			rhp.CorsConfig.AccessControlAllowMethods.SetQuantity(int64(len(rhp.CorsConfig.AccessControlAllowMethods.Items)))
		}
		if rhp.CorsConfig.AccessControlAllowOrigins != nil && rhp.CorsConfig.AccessControlAllowOrigins.Items != nil {
			rhp.CorsConfig.AccessControlAllowOrigins.SetQuantity(int64(len(rhp.CorsConfig.AccessControlAllowOrigins.Items)))
		}
		if rhp.CorsConfig.AccessControlExposeHeaders != nil && rhp.CorsConfig.AccessControlExposeHeaders.Items != nil {
			rhp.CorsConfig.AccessControlExposeHeaders.SetQuantity(int64(len(rhp.CorsConfig.AccessControlExposeHeaders.Items)))
		}
	}
	if rhp.CustomHeadersConfig != nil && rhp.CustomHeadersConfig.Items != nil {
		rhp.CustomHeadersConfig.SetQuantity(int64(len(rhp.CustomHeadersConfig.Items)))
	}
	if rhp.RemoveHeadersConfig != nil && rhp.RemoveHeadersConfig.Items != nil {
		rhp.RemoveHeadersConfig.SetQuantity(int64(len(rhp.RemoveHeadersConfig.Items)))
	}
}
