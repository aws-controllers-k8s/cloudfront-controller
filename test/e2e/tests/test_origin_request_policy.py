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

"""Integration tests for the CloudFront OriginRequestPolicy resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import origin_request_policy

ORIGIN_REQUEST_POLICY_RESOURCE_PLURAL = "originrequestpolicies"
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10
MIN_TTL = 600


@pytest.fixture(scope="module")
def simple_origin_request_policy():
    origin_request_policy_name = random_suffix_name("my-orp", 24)
    origin_request_policy_comment = "a simple origin_request_policy"

    replacements = REPLACEMENT_VALUES.copy()
    replacements['ORIGIN_REQUEST_POLICY_NAME'] = origin_request_policy_name
    replacements['ORIGIN_REQUEST_POLICY_COMMENT'] = origin_request_policy_comment

    resource_data = load_resource(
        "origin_request_policy",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, ORIGIN_REQUEST_POLICY_RESOURCE_PLURAL,
        origin_request_policy_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'id' in cr['status']
    origin_request_policy_id = cr['status']['id']

    origin_request_policy.wait_until_exists(origin_request_policy_id)

    yield (ref, cr, origin_request_policy_id)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    origin_request_policy.wait_until_deleted(origin_request_policy_id)


@service_marker
@pytest.mark.canary
class TestOriginRequestPolicy:
    def test_crud(self, simple_origin_request_policy):
        ref, res, origin_request_policy_id = simple_origin_request_policy

        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'originRequestPolicyConfig' in cr['spec']
        assert 'comment' in cr['spec']['originRequestPolicyConfig']
        assert cr['spec']['originRequestPolicyConfig']['comment'] == 'a simple origin_request_policy'

        condition.assert_ready(ref)

        latest = origin_request_policy.get(origin_request_policy_id)
        assert latest is not None
        assert 'OriginRequestPolicyConfig' in latest
        assert 'Comment' in latest['OriginRequestPolicyConfig']
        assert latest['OriginRequestPolicyConfig']['Comment'] == 'a simple origin_request_policy'

        updates = {
            "spec": {
                "originRequestPolicyConfig": {
                    "comment": 'new comment'
                }
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = origin_request_policy.get(origin_request_policy_id)
        assert latest is not None
        assert 'OriginRequestPolicyConfig' in latest
        assert 'Comment' in latest['OriginRequestPolicyConfig']
        assert latest['OriginRequestPolicyConfig']['Comment'] == 'new comment'
