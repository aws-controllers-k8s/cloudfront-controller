	if functionAutoPublishEnabled(ko) {
		if err := rm.publishFunction(ctx, ko); err != nil {
			return &resource{ko}, err
		}
	}
