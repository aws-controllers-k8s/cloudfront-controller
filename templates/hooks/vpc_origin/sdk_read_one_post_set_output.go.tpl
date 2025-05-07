// We need to do this manually because the ETag field (required for the
// IfMatch field when updating or deleting a resource in CloudFront) is
// outside the wrapper field of `ResponseHeadersPolicy` in the response
// and therefore not picked up by SetResource code generation.
if resp.ETag != nil {
	ko.Status.ETag = resp.ETag
}

// We need to get the tags that are in the AWS resource
ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))