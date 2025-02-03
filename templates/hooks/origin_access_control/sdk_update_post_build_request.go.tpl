	// If we don't do this, we get the following on every update call:
	// InvalidIfMatchVersion: The If-Match version is missing or not valid for the resource.
	if latest.ko.Status.ETag != nil {
		input.IfMatch =	latest.ko.Status.ETag
	}