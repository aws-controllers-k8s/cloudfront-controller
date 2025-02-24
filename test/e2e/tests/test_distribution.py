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

"""Integration tests for the CloudFront Distribution resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.aws import identity
from acktest import tags
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import distribution

DISTRIBUTION_RESOURCE_PLURAL = "distributions"
DELETE_WAIT_AFTER_SECONDS = 120
CHECK_STATUS_WAIT_SECONDS = 300
MODIFY_WAIT_AFTER_SECONDS = 300


@pytest.fixture(scope="module")
def simple_distribution():
    distribution_name = random_suffix_name("my-distribution", 24)
    origin_id = random_suffix_name("origin", 12)
    distribution_comment = "a simple distribution"

    region = identity.get_region()
    resources = get_bootstrap_resources()
    bucket_name = resources.PublicBucket.name
    bucket_domain_name = f"{bucket_name}.s3.amazonaws.com"

    replacements = REPLACEMENT_VALUES.copy()
    replacements['DISTRIBUTION_NAME'] = distribution_name
    replacements['DISTRIBUTION_COMMENT'] = distribution_comment
    replacements['ORIGIN_ID'] = origin_id
    replacements['ORIGIN_S3_DOMAIN_NAME'] = bucket_domain_name

    resource_data = load_resource(
        "distribution",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, DISTRIBUTION_RESOURCE_PLURAL,
        distribution_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'id' in cr['status']
    distribution_id = cr['status']['id']

    distribution.wait_until_exists(distribution_id)

    yield (ref, cr, distribution_id)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted
    distribution.wait_until_deleted(distribution_id)


@service_marker
@pytest.mark.canary
class TestDistribution:
    def test_crud(self, simple_distribution):
        ref, res, distribution_id = simple_distribution

        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        # Before we update the Distribution CR below, let's check to see that
        # the MinTTL field in the CR is still what we set in the original
        # Create call.
        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'distributionConfig' in cr['spec']
        assert 'enabled' in cr['spec']['distributionConfig']
        assert bool(cr['spec']['distributionConfig']['enabled']) == True

        assert k8s.wait_on_condition(
            ref,
            "ACK.ResourceSynced",
            "True",
            wait_periods=CHECK_STATUS_WAIT_SECONDS // 10,
        )

        latest = distribution.get(distribution_id)
        assert latest is not None
        assert 'DistributionConfig' in latest
        assert 'Enabled' in latest['DistributionConfig']
        assert bool(latest['DistributionConfig']['Enabled']) == True
        
        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']

        assert 'tags' in cr['spec']
        user_tags = cr['spec']['tags']

        response_tags = distribution.get_tags(arn)

        tags.assert_ack_system_tags(
            tags=response_tags,
        )

        user_tags = [{"Key": d["key"], "Value": d["value"]} for d in user_tags]
        tags.assert_equal_without_ack_tags(
            expected=user_tags,
            actual=response_tags,
        )

        # We're now going to modify the enabled field of the Distribution, wait
        # some time and verify that the CloudFront server-side resource shows
        # the new value of the field.
        #
        # NOTE(jaypipes): It is necessary to disable the Distribution before
        # deleting it...
        updates = {
            "spec": {
                "distributionConfig": {
                    "enabled": False
                },
                "tags": [
                    {
                        "key": "another",
                        "value": "here"
                    }
                ]
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = distribution.get(distribution_id)
        assert latest is not None
        assert 'DistributionConfig' in latest
        assert 'Enabled' in latest['DistributionConfig']
        assert bool(latest['DistributionConfig']['Enabled']) == False

        cr = k8s.get_resource(ref)

        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']

        assert 'tags' in cr['spec']
        user_tags = cr['spec']['tags']

        response_tags = distribution.get_tags(arn)

        tags.assert_ack_system_tags(
            tags=response_tags,
        )

        user_tags = [{"Key": d["key"], "Value": d["value"]} for d in user_tags]
        tags.assert_equal_without_ack_tags(
            expected=user_tags,
            actual=response_tags,
        )
