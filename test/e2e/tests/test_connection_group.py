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

"""Integration tests for the CloudFront ConnectionGroup resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest import tags
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import connection_group

CONNECTION_GROUP_RESOURCE_PLURAL = "connectiongroups"
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 60
CHECK_STATUS_WAIT_PERIODS = 20
MODIFY_WAIT_AFTER_SECONDS = 10


@pytest.fixture(scope="module")
def simple_connection_group():
    connection_group_name = random_suffix_name("cloudfront-test-cg", 24)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CONNECTION_GROUP_NAME"] = connection_group_name
    replacements["CONNECTION_GROUP_ENABLED"] = "true"
    replacements["CONNECTION_GROUP_IPV6_ENABLED"] = "true"

    resource_data = load_resource(
        "connection_group",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, CONNECTION_GROUP_RESOURCE_PLURAL,
        connection_group_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'id' in cr['status']
    connection_group_id = cr['status']['id']

    connection_group.wait_until_exists(connection_group_id)

    yield (ref, cr, connection_group_id)

    # A connection group must be disabled before it can be deleted, otherwise
    # the CloudFront API returns a ResourceNotDisabled error. Disable it here so
    # teardown succeeds regardless of the test outcome (idempotent if the test
    # body already disabled it).
    k8s.patch_custom_resource(ref, {"spec": {"enabled": False}})
    time.sleep(MODIFY_WAIT_AFTER_SECONDS)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    connection_group.wait_until_deleted(connection_group_id)


@service_marker
@pytest.mark.canary
class TestConnectionGroup:
    def test_crud(self, simple_connection_group):
        ref, _, connection_group_id = simple_connection_group

        assert k8s.wait_on_condition(
            ref,
            "ACK.ResourceSynced",
            "True",
            wait_periods=CHECK_STATUS_WAIT_PERIODS,
            period_length=CHECK_STATUS_WAIT_SECONDS
        )

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'name' in cr['spec']
        assert cr['spec']['enabled'] is True
        assert cr['spec']['ipv6Enabled'] is True

        # Verify the resource state in the CloudFront API matches the spec,
        # including the tags applied at create time.
        latest = connection_group.get(connection_group_id)
        assert latest is not None
        assert latest['Enabled'] is True
        assert latest['Ipv6Enabled'] is True

        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']

        assert 'tags' in cr['spec']
        user_tags = [{"Key": t["key"], "Value": t["value"]} for t in cr['spec']['tags']]

        response_tags = connection_group.get_tags(arn)
        tags.assert_ack_system_tags(
            tags=response_tags,
        )

        tags.assert_equal_without_ack_tags(
            expected=user_tags,
            actual=response_tags,
        )

        # Test update: disable the connection group and change its tags.
        updates = {
            "spec": {
                "enabled": False,
                "tags": [
                    {
                        "key": "new_key",
                        "value": "new_value"
                    }
                ]
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = connection_group.get(connection_group_id)
        assert latest is not None
        assert latest['Enabled'] is False

        response_tags = connection_group.get_tags(arn)
        assert {"Key": "new_key", "Value": "new_value"} in response_tags
        assert {"Key": "hello", "Value": "world"} not in response_tags

        # Test: Name field is immutable and rejected by the API server
        with pytest.raises(k8s.ApiException) as exc:
            k8s.patch_custom_resource(ref, {"spec": {"name": "new_cg_name"}})
        assert "Value is immutable once set" in str(exc.value.body)