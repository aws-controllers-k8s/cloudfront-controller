	if resp.FunctionSummary != nil {
		if resp.FunctionSummary.FunctionMetadata != nil {
			if resp.FunctionSummary.FunctionMetadata.FunctionARN != nil {
				ko.Status.ACKResourceMetadata.ARN = (*ackv1alpha1.AWSResourceName)(resp.FunctionSummary.FunctionMetadata.FunctionARN)
			}
			if resp.FunctionSummary.FunctionConfig != nil {
				if resp.FunctionSummary.FunctionConfig.Runtime != "" {
					ko.Spec.FunctionConfig.Runtime = aws.String(string(resp.FunctionSummary.FunctionConfig.Runtime))
				}
				if resp.FunctionSummary.FunctionConfig.Comment != nil {
					ko.Spec.FunctionConfig.Comment = resp.FunctionSummary.FunctionConfig.Comment
				}
			}
		}
	}
	if err := rm.setResourceAdditionalFields(ctx, ko); err != nil {
		return nil, err
	}
