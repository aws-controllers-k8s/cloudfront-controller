	if functionAutoPublishEnabled(ko) {
		msg := "function created successfully. Requeue to autopublish"
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, &msg, nil)
	}
