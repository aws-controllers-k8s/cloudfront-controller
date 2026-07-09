	// CloudFront rejects deleting an enabled connection group with a
	// ResourceNotDisabled error. If the connection group is still enabled,
	// disable it first via an update, then requeue so the delete can proceed
	// once the disable has taken effect.
	if r.ko.Spec.Enabled != nil && *r.ko.Spec.Enabled {
		r.ko.Spec.Enabled = aws.Bool(false)
		// An empty path here works because it is a non-Tags change, 
		// forcing an update which then disables the resource without calling syncTags
		_, err = rm.sdkUpdate(ctx, r, r, &ackcompare.Delta{
			Differences: []*ackcompare.Difference{&ackcompare.Difference{Path: ackcompare.Path{}, A: nil, B: nil}},
		})
		if err != nil {
			return r, err
		}
		return r, ackrequeue.NeededAfter(
			fmt.Errorf("waiting for connection group to be disabled before deletion"),
			ackrequeue.DefaultRequeueAfterDuration,
		)
	}

