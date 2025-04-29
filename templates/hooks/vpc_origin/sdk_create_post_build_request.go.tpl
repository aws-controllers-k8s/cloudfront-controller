// CloudFront's API uses XML with nested list structures containing 'items' and
// 'quantity' fields. To avoid InconsistentQuantities errors (where quantity must
// match the number of items), we'll derive the quantity dynamically using
// len(items) instead of maintaining it separately.
setQuantityFields(input.VpcOriginEndpointConfig)