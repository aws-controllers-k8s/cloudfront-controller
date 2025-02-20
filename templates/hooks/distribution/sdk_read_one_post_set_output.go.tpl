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
	// We need to get the tags that are in the AWS resource
	ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
	if !distributionDeployed(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	} else {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
	}