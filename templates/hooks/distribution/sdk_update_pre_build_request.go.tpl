	// return desired with latest status
	updatedDesired := desired.DeepCopy()
	updatedDesired.SetStatus(latest)
	if delta.DifferentAt("Spec.Tags") {
		err := rm.syncTags(
			ctx,
			latest,
			desired,
		)
		if err != nil {
			return nil, err
		}
	}
	if !delta.DifferentExcept("Spec.Tags") {
		return rm.concreteResource(updatedDesired), nil
	}

	if !distributionDeployed(latest) {
		msg := "Distribution is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(updatedDesired, corev1.ConditionFalse, &msg, nil)
		return rm.concreteResource(updatedDesired), requeueWaitUntilCanModify(latest)
	}
