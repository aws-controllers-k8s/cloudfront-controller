	// We should re-use the same token we used to create the Distribution resource.
	if latest.ko.Status.CallerReference != nil {
		input.DistributionConfig.SetCallerReference(*latest.ko.Status.CallerReference)
	}
	// If we don't do this, we get the following on every update call:
	// InvalidIfMatchVersion: The If-Match version is missing or not valid for the resource.
	if latest.ko.Status.ETag != nil {
		input.SetIfMatch(*latest.ko.Status.ETag)
	}
	// ¯\\\_(ツ)_/¯
	if input.DistributionConfig != nil {
		setQuantityFields(input.DistributionConfig)
	}
