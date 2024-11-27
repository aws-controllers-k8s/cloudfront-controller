	if r.ko.Spec.DistributionConfig.Enabled != nil && *r.ko.Spec.DistributionConfig.Enabled {
		resourceCopy := r.ko.DeepCopy()
		// If the distribution is enabled make sure to disable it before making a delete API call
		resourceCopy.Spec.DistributionConfig.Enabled = aws.Bool(false)
		_, err := rm.sdkUpdate(ctx, &resource{resourceCopy}, r, nil)
		if err != nil {
			return nil, err
		}

		// Now keep requeing until the distribution is disabled.
		return nil, requeueWaitInProgress
	}