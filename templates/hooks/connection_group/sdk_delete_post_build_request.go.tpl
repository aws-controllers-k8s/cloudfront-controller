// If we don't do this, we get the following on every delete call:
// InvalidIfMatchVersion: The If-Match version is missing or not valid for the resource.
if r.ko.Status.ETag != nil {
	input.IfMatch = r.ko.Status.ETag
}
