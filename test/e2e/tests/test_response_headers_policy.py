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

"""Integration tests for the CloudFront ResponseHeadersPolicy resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import response_headers_policy

RESPONSE_HEADERS_POLICY_RESOURCE_PLURAL = "responseheaderspolicies"
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10
MIN_TTL = 600


@pytest.fixture(scope="module")
def simple_response_headers_policy():
    response_headers_policy_name = random_suffix_name("my-rhp", 24)
    response_headers_policy_comment = "a simple response_headers_policy"

    replacements = REPLACEMENT_VALUES.copy()
    replacements['RESPONSE_HEADERS_POLICY_NAME'] = response_headers_policy_name
    replacements['RESPONSE_HEADERS_POLICY_COMMENT'] = response_headers_policy_comment

    resource_data = load_resource(
        "response_headers_policy",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESPONSE_HEADERS_POLICY_RESOURCE_PLURAL,
        response_headers_policy_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'id' in cr['status']
    response_headers_policy_id = cr['status']['id']

    response_headers_policy.wait_until_exists(response_headers_policy_id)

    yield (ref, cr, response_headers_policy_id)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    response_headers_policy.wait_until_deleted(response_headers_policy_id)


@service_marker
@pytest.mark.canary
class TestResponseHeadersPolicy:
    def test_crud(self, simple_response_headers_policy):
        ref, res, response_headers_policy_id = simple_response_headers_policy

        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'responseHeadersPolicyConfig' in cr['spec']
        assert 'comment' in cr['spec']['responseHeadersPolicyConfig']
        assert cr['spec']['responseHeadersPolicyConfig']['comment'] == 'a simple response_headers_policy'

        condition.assert_ready(ref)

        latest = response_headers_policy.get(response_headers_policy_id)
        assert latest is not None
        assert 'ResponseHeadersPolicyConfig' in latest
        assert 'Comment' in latest['ResponseHeadersPolicyConfig']
        assert latest['ResponseHeadersPolicyConfig']['Comment'] == 'a simple response_headers_policy'

        updates = {
            "spec": {
                "responseHeadersPolicyConfig": {
                    "comment": 'new comment',
                    "customHeadersConfig": {
                        "items": [
                            {
                                "header": "X-ACK-CONTROLLER",
                                "override": True,
                                "value": "CLOUDFRONT"
                            },
                            {
                                "header": "X-ACK-RESOURCE",
                                "override": True,
                                "value": "RESPONSE_HEADERS_POLICY"
                            }
                        ]
                    }
                }
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = response_headers_policy.get(response_headers_policy_id)
        assert latest is not None
        assert 'ResponseHeadersPolicyConfig' in latest
        assert 'Comment' in latest['ResponseHeadersPolicyConfig']
        assert latest['ResponseHeadersPolicyConfig']['Comment'] == 'new comment'
        assert 'CustomHeadersConfig' in latest['ResponseHeadersPolicyConfig']
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Quantity'] == 2
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Items'][0]['Header'] == 'X-ACK-CONTROLLER'
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Items'][0]['Value'] == 'CLOUDFRONT'
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Items'][0]['Override'] == True
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Items'][1]['Header'] == 'X-ACK-RESOURCE'
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Items'][1]['Value'] == 'RESPONSE_HEADERS_POLICY'
        assert latest['ResponseHeadersPolicyConfig']['CustomHeadersConfig']['Items'][1]['Override'] == True
