	if !distributionDeployed(r) {
		msg := "Distribution is in '" + *r.ko.Status.Status + "' status"
		ackcondition.SetSynced(r, corev1.ConditionFalse, &msg, nil)
		return r, requeueWaitUntilCanModify(r)
	}
	if r.ko.Spec.DistributionConfig.Enabled != nil && *r.ko.Spec.DistributionConfig.Enabled {
		r.ko.Spec.DistributionConfig.Enabled = aws.Bool(false)
		_, err = rm.sdkUpdate(ctx, r, r, &ackcompare.Delta{
			Differences: []*ackcompare.Difference{&ackcompare.Difference{Path: ackcompare.Path{}, A: nil, B: nil}, },
		})
		return r, ackrequeue.NeededAfter(
			fmt.Errorf("waiting for distribution to be disabled before deletion"),
			ackrequeue.DefaultRequeueAfterDuration,
		)
	}
