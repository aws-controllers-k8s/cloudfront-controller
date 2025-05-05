// CloudFront's API uses XML with nested list structures containing 'items' and
// 'quantity' fields. To avoid InconsistentQuantities errors (where quantity must
// match the number of items), we'll derive the quantity dynamically using
// len(items) instead of maintaining it separately.
setQuantityFields(input.VpcOriginEndpointConfig)

// CloudFront supports applying tags on create. However, the code generator
// doesn't quite match the shape of the CRD spec to the CreateVpcOriginInput.
updateTagsInCreateRequest(desired, input)