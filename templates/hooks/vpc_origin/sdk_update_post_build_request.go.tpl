// If we don't do this, we get the following on every update call:
// InvalidIfMatchVersion: The If-Match version is missing or not valid for the resource.
if latest.ko.Status.ETag != nil {
	input.IfMatch = latest.ko.Status.ETag
}

// CloudFront's API uses XML with nested list structures containing 'items' and
// 'quantity' fields. To avoid InconsistentQuantities errors (where quantity must
// match the number of items), we'll derive the quantity dynamically using
// len(items) instead of maintaining it separately.
setQuantityFields(input.VpcOriginEndpointConfig)