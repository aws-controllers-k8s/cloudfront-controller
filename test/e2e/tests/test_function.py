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

"""Integration tests for the CloudFront Function resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.aws import identity
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import function

FUNCTIONS_RESOURCE_PLURAL = "functions"
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 30
MODIFY_WAIT_AFTER_SECONDS = 30


@pytest.fixture(scope="module")
def simple_function():
    function_name = random_suffix_name("my-function", 24)

    # resources = get_bootstrap_resources()
    # TODO(a-hilaly) replace example.com with bucket domain name
    # bucket_name = resources.PublicBucket.name
    # bucket_domain_name = f"{bucket_name}.s3.amazonaws.com"

    replacements = REPLACEMENT_VALUES.copy()
    replacements['FUNCTION_NAME'] = function_name
    replacements['FUNCTION_RUNTIME'] = "cloudfront-js-2.0"
    replacements['DOMAIN_NAME'] = "example.com"

    resource_data = load_resource(
        "function",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, FUNCTIONS_RESOURCE_PLURAL,
        function_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'eTag' in cr['status']

    function.wait_until_exists(function_name)

    yield (ref, cr, function_name)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    function.wait_until_deleted(function_name)


@service_marker
@pytest.mark.canary
class TestFunction:
    def test_crud(self, simple_function):
        ref, res, function_name = simple_function

        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        # Check that function exists
        cr = k8s.get_resource(ref)
        assert cr is not None

        condition.assert_synced(ref)

        latest = function.get(function_name)
        assert latest is not None
        assert latest['Status'] == "UNPUBLISHED"

        # Update the function
        updates = {
            "spec": {
                "functionConfig": {
                    "comment": "New comment"
                }
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = function.get(function_name)
        assert latest is not None
        assert latest['FunctionConfig']['Comment'] == "New comment"
