// CloudFront supports applying tags on create. However, the code generator
// doesn't quite match the shape of the CRD spec to the
// CreateDistributionTenantInput, so we set them manually here.
updateTagsInCreateRequest(desired, input)
