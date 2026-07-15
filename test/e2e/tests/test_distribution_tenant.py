# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the CloudFront DistributionTenant resource"""

import pytest

from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import get_bootstrap_resources

DISTRIBUTION_TENANT_RESOURCE_PLURAL = "distributiontenants"


@pytest.fixture(scope="module")
def simple_distribution_tenant():
    resources = get_bootstrap_resources()
    dist_id = resources.TenantDistribution.distribution_id
    cg_id = resources.TenantConnectionGroup.connection_group_id

    tenant_name = random_suffix_name("dt-test-tenant", 24)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["DISTRIBUTION_TENANT_NAME"] = tenant_name
    replacements["DISTRIBUTION_ID"] = dist_id
    replacements["CONNECTION_GROUP_ID"] = cg_id
    replacements["DISTRIBUTION_TENANT_DOMAIN"] = "ack.test.example.com"

    resource_data = load_resource(
        "distribution_tenant",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, DISTRIBUTION_TENANT_RESOURCE_PLURAL,
        tenant_name, namespace="default",
    )

    k8s.create_custom_resource(ref, resource_data)
    k8s.wait_resource_consumed_by_controller(ref)

    yield ref

    _, deleted = k8s.delete_custom_resource(ref)
    assert deleted


@service_marker
@pytest.mark.canary
class TestDistributionTenant:
    def test_create_fails_invalid_domain(self, simple_distribution_tenant):
        """DistributionTenants require a valid domain
        with proper DNS and certificate configuration. Because the e2e
        test environment does not have a domain available, we can only
        verify that the controller correctly surfaces the InvalidArgument
        error from the CloudFront API as a Terminal condition when an
        invalid domain is provided.
        """
        ref = simple_distribution_tenant

        condition.assert_type_status(
            ref, condition.CONDITION_TYPE_TERMINAL,
        )

        terminal_condition = k8s.get_resource_condition(
            ref, condition.CONDITION_TYPE_TERMINAL,
        )
        assert terminal_condition is not None
        assert terminal_condition['status'] == "True"

        assert "InvalidArgument" in terminal_condition['message']
