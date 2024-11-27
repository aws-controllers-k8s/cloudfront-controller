    // We need to do this manually because the ETag field (required for the
    // IfMatch field when updating or deleting a resource in CloudFront) is
    // outside the wrapper field of `Distribution` in the response and
    // therefore not picked up by SetResource code generation.
	if resp.ETag != nil {
		ko.Status.ETag = resp.ETag
	}
	// We need to set the CallerReference here. All the Update operations
	// will have to re-use the same CallerReference.
	if resp.Distribution != nil &&  resp.Distribution.DistributionConfig != nil {
		ko.Status.CallerReference = resp.Distribution.DistributionConfig.CallerReference
	}

	// Requeue if the distribution is in progress
	if distributionInProgress(&resource{ko}) {
		return &resource{ko}, requeueWaitInProgress
	}