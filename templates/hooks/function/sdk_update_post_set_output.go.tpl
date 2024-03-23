	if functionAutoPublishEnabled(ko) {
		if err := rm.publishFunction(ctx, ko); err != nil {
			return nil, err
		}
	}