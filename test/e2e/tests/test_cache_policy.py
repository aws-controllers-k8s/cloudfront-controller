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

"""Integration tests for the CloudFront CachePolicy resource"""

import time

import pytest

from acktest.k8s import condition
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e import cache_policy

CACHE_POLICY_RESOURCE_PLURAL = "cachepolicies"
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10
MIN_TTL = 600


@pytest.fixture(scope="module")
def simple_cache_policy():
    cache_policy_name = random_suffix_name("my-cache-policy", 24)
    cache_policy_comment = "a simple cache_policy"

    replacements = REPLACEMENT_VALUES.copy()
    replacements['CACHE_POLICY_NAME'] = cache_policy_name
    replacements['CACHE_POLICY_COMMENT'] = cache_policy_comment
    replacements['MIN_TTL'] = str(MIN_TTL)

    resource_data = load_resource(
        "cache_policy",
        additional_replacements=replacements,
    )

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, CACHE_POLICY_RESOURCE_PLURAL,
        cache_policy_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert k8s.get_resource_exists(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'id' in cr['status']
    cache_policy_id = cr['status']['id']

    cache_policy.wait_until_exists(cache_policy_id)

    yield (ref, cr, cache_policy_id)

    _, deleted = k8s.delete_custom_resource(
        ref,
        period_length=DELETE_WAIT_AFTER_SECONDS,
    )
    assert deleted

    cache_policy.wait_until_deleted(cache_policy_id)


@service_marker
@pytest.mark.canary
class TestCachePolicy:
    def test_crud(self, simple_cache_policy):
        ref, res, cache_policy_id = simple_cache_policy

        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        # Before we update the CachePolicy CR below, let's check to see that
        # the MinTTL field in the CR is still what we set in the original
        # Create call.
        cr = k8s.get_resource(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'cachePolicyConfig' in cr['spec']
        assert 'minTTL' in cr['spec']['cachePolicyConfig']
        assert cr['spec']['cachePolicyConfig']['minTTL'] == MIN_TTL

        condition.assert_synced(ref)

        latest = cache_policy.get(cache_policy_id)
        assert latest is not None
        assert 'CachePolicyConfig' in latest
        assert 'MinTTL' in latest['CachePolicyConfig']
        assert latest['CachePolicyConfig']['MinTTL'] == MIN_TTL

        new_min_ttl = MIN_TTL + 100

        # We're now going to modify the MinTTL field of the CachePolicy, wait
        # some time and verify that the CloudFront server-side resource shows
        # the new value of the field.
        updates = {
            "spec": {
                "cachePolicyConfig": {
                    "minTTL": new_min_ttl
                }
            },
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        latest = cache_policy.get(cache_policy_id)
        assert latest is not None
        assert 'CachePolicyConfig' in latest
        assert 'MinTTL' in latest['CachePolicyConfig']
        assert latest['CachePolicyConfig']['MinTTL'] == new_min_ttl
