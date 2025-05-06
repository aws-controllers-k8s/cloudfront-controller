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

"""Integration tests for the CloudFront VpcOrigin resource"""

import time
import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e import vpc_origin
from logging import getLogger

VPC_ORIGIN_RESOURCE_PLURAL = "vpcorigins"
DELETE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_PERIODS = 3
# VPC Origins can take up to 15 minutes to create :(
CHECK_STATUS_WAIT_PERIODS = 20
CHECK_STATUS_WAIT_SECONDS = 60
MODIFY_WAIT_AFTER_SECONDS = 30

logger = getLogger(__name__)


@pytest.fixture(scope="module")
def simple_vpc_origin():
    vpc_origin_name = random_suffix_name("cloudfront-test-vpc-origin", 32)
    vpc_origin_protocol_policy = "http-only"
    vpc_origin_ssl_protocols_1 = "TLSv1.2"
    vpc_origin_http_port = "80"
    vpc_origin_https_port = "443"
    vpc_origin_tag_key = "tag1"
    vpc_origin_tag_value = "value1"

    replacements = REPLACEMENT_VALUES.copy()
    replacements["VPC_ORIGIN_NAME"] = vpc_origin_name
    replacements["VPC_ORIGIN_ENDPOINT_ARN"] = get_bootstrap_resources().NetworkLoadBalancer.arn
    replacements["VPC_ORIGIN_HTTP_PORT"] = vpc_origin_http_port
    replacements["VPC_ORIGIN_HTTPS_PORT"] = vpc_origin_https_port
    replacements["VPC_ORIGIN_PROTOCOL_POLICY"] = vpc_origin_protocol_policy
    replacements["SSL_PROTOCOL_1"] = vpc_origin_ssl_protocols_1
    replacements["TAG_KEY"] = vpc_origin_tag_key
    replacements["TAG_VALUE"] = vpc_origin_tag_value

    resource_data = load_resource(
        "vpc_origin",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        VPC_ORIGIN_RESOURCE_PLURAL,
        vpc_origin_name,
        namespace="default",
    )

    logger.info("Creating VPCOrigin %s", vpc_origin_name)
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert "status" in cr
    assert "id" in cr["status"]
    vpc_origin_id = cr["status"]["id"]

    vpc_origin.wait_until_exists(vpc_origin_id)
    logger.info("VPCOrigin %s exists in cluster", vpc_origin_name)

    yield (ref, cr, vpc_origin_id)

    logger.info("Deleting VPCOrigin %s", vpc_origin_name)
    _, deleted = k8s.delete_custom_resource(
        ref,
        wait_periods=DELETE_WAIT_PERIODS,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    vpc_origin.wait_until_deleted(vpc_origin_id)


@service_marker
@pytest.mark.canary
class TestVpcOrigin:
    def test_crud(self, simple_vpc_origin):
        ref, res, vpc_origin_id = simple_vpc_origin

        assert k8s.wait_on_condition(
            ref,
            "ACK.ResourceSynced",
            "True",
            wait_periods=CHECK_STATUS_WAIT_PERIODS,
            period_length=CHECK_STATUS_WAIT_SECONDS
        )

        cr = k8s.get_resource(ref)
        assert cr is not None
        assert "spec" in cr
        assert "vpcOriginEndpointConfig" in cr["spec"]
        assert "originProtocolPolicy" in cr["spec"]["vpcOriginEndpointConfig"]
        assert cr["spec"]["vpcOriginEndpointConfig"]["originProtocolPolicy"] == "http-only"

        latest = vpc_origin.get(vpc_origin_id)
        assert latest is not None
        assert "VpcOriginEndpointConfig" in latest
        assert "OriginProtocolPolicy" in latest["VpcOriginEndpointConfig"]
        assert latest["VpcOriginEndpointConfig"]["OriginProtocolPolicy"] == "http-only"

        # Test update
        updates = {
            "spec": {"vpcOriginEndpointConfig": {"originProtocolPolicy": "https-only"}},
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = vpc_origin.get(vpc_origin_id)
        assert latest is not None
        assert "VpcOriginEndpointConfig" in latest
        assert "OriginProtocolPolicy" in latest["VpcOriginEndpointConfig"]
        assert latest["VpcOriginEndpointConfig"]["OriginProtocolPolicy"] == "https-only"