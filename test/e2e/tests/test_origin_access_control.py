# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the CloudFront OriginAccessControl resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import origin_access_control

ORIGIN_ACCESS_CONTROL_RESOURCE_PLURAL = "originaccesscontrols"
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10


@pytest.fixture(scope="module")
def simple_origin_access_control():
    origin_access_control_name = random_suffix_name("my-oac", 24)
    origin_access_control_description = "a simple origin_access_control"

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ORIGIN_ACCESS_CONTROL_NAME"] = origin_access_control_name
    replacements["ORIGIN_ACCESS_CONTROL_DESCRIPTION"] = (
        origin_access_control_description
    )

    resource_data = load_resource(
        "origin_access_control",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        ORIGIN_ACCESS_CONTROL_RESOURCE_PLURAL,
        origin_access_control_name,
        namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert "status" in cr
    assert "id" in cr["status"]
    origin_access_control_id = cr["status"]["id"]

    origin_access_control.wait_until_exists(origin_access_control_id)

    yield (ref, cr, origin_access_control_id)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    origin_access_control.wait_until_deleted(origin_access_control_id)


@service_marker
@pytest.mark.canary
class TestOriginAccessControl:
    def test_crud(self, simple_origin_access_control):
        ref, res, origin_access_control_id = simple_origin_access_control

        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert "spec" in cr
        assert "originAccessControlConfig" in cr["spec"]
        assert "description" in cr["spec"]["originAccessControlConfig"]
        assert (
            cr["spec"]["originAccessControlConfig"]["description"]
            == "a simple origin_access_control"
        )

        condition.assert_ready(ref)

        latest = origin_access_control.get(origin_access_control_id)
        assert latest is not None
        assert "OriginAccessControlConfig" in latest
        assert "Description" in latest["OriginAccessControlConfig"]
        assert (
            latest["OriginAccessControlConfig"]["Description"]
            == "a simple origin_access_control"
        )

        updates = {
            "spec": {"originAccessControlConfig": {"description": "new description"}},
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = origin_access_control.get(origin_access_control_id)
        assert latest is not None
        assert "OriginAccessControlConfig" in latest
        assert "Description" in latest["OriginAccessControlConfig"]
        assert latest["OriginAccessControlConfig"]["Description"] == "new description"
