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

"""Bootstrappable CloudFront resources for distribution tenant e2e tests.

These resources are created once during bootstrap and reused across test runs,
avoiding the 10+ minute setup/teardown for connection groups and multi-tenant
distributions on every test invocation.
"""

import logging
import time
import boto3

from dataclasses import dataclass, field
from acktest.bootstrapping import Bootstrappable
from acktest import resources


WAIT_INTERVAL_SECONDS = 15
WAIT_TIMEOUT_SECONDS = 60 * 15


def _wait_until_deployed(describe_fn, resource_type):
    """Polls a describe function until the resource status is 'Deployed'."""
    deadline = time.time() + WAIT_TIMEOUT_SECONDS
    while time.time() < deadline:
        resp = describe_fn()
        status = resp.get("Status", "")
        if status == "Deployed":
            logging.info(f"{resource_type} reached Deployed status")
            return resp
        logging.info(f"Waiting for {resource_type} to deploy (status: {status})")
        time.sleep(WAIT_INTERVAL_SECONDS)
    raise Exception(f"Timed out waiting for {resource_type} to reach Deployed status")


@dataclass
class ConnectionGroup(Bootstrappable):
    # Inputs
    name_prefix: str

    # Outputs
    connection_group_id: str = field(init=False, default=None)
    arn: str = field(init=False, default=None)

    @property
    def cf_client(self):
        return boto3.client("cloudfront")

    def bootstrap(self):
        super().bootstrap()

        self.name = resources.random_suffix_name(self.name_prefix, 63)
        logging.info(f"Creating ConnectionGroup: {self.name}")

        resp = self.cf_client.create_connection_group(
            Name=self.name,
            Enabled=True,
            Ipv6Enabled=True,
        )

        cg = resp["ConnectionGroup"]
        self.connection_group_id = cg["Id"]
        self.arn = cg["Arn"]

        _wait_until_deployed(
            lambda: self.cf_client.get_connection_group(
                Identifier=self.connection_group_id
            )["ConnectionGroup"],
            f"ConnectionGroup {self.name}",
        )

        logging.info(
            f"ConnectionGroup bootstrapped: id={self.connection_group_id}"
        )

    def cleanup(self):
        if self.connection_group_id:
            logging.info(
                f"Cleaning up ConnectionGroup: {self.connection_group_id}"
            )
            try:
                # Disable before delete
                self.cf_client.update_connection_group(
                    Identifier=self.connection_group_id,
                    Enabled=False,
                    IfMatch="*",
                )
                _wait_until_deployed(
                    lambda: self.cf_client.get_connection_group(
                        Identifier=self.connection_group_id
                    )["ConnectionGroup"],
                    f"ConnectionGroup {self.connection_group_id} (disable)",
                )
                self.cf_client.delete_connection_group(
                    Identifier=self.connection_group_id,
                    IfMatch="*",
                )
            except Exception as ex:
                logging.error(
                    f"Error cleaning up ConnectionGroup "
                    f"{self.connection_group_id}: {ex}"
                )

        super().cleanup()


@dataclass
class MultiTenantDistribution(Bootstrappable):
    # Inputs
    name_prefix: str
    bucket_domain_name: str

    # Outputs
    distribution_id: str = field(init=False, default=None)
    domain_name: str = field(init=False, default=None)
    arn: str = field(init=False, default=None)

    @property
    def cf_client(self):
        return boto3.client("cloudfront")

    def _get_caching_optimized_policy_id(self):
        resp = self.cf_client.list_cache_policies(Type="managed")
        for item in resp["CachePolicyList"]["Items"]:
            if item["CachePolicy"]["CachePolicyConfig"]["Name"] == "Managed-CachingOptimized":
                return item["CachePolicy"]["Id"]
        raise Exception("Managed-CachingOptimized cache policy not found")

    def bootstrap(self):
        super().bootstrap()

        self.name = resources.random_suffix_name(self.name_prefix, 63)
        origin_id = resources.random_suffix_name("origin", 32)
        cache_policy_id = self._get_caching_optimized_policy_id()

        logging.info(f"Creating multi-tenant Distribution: {self.name}")

        resp = self.cf_client.create_distribution(
            DistributionConfig={
                "CallerReference": self.name,
                "Comment": "Bootstrap multi-tenant distribution for tenant tests",
                "Enabled": True,
                "ConnectionMode": "tenant-only",
                "DefaultCacheBehavior": {
                    "TargetOriginId": origin_id,
                    "ViewerProtocolPolicy": "allow-all",
                    "CachePolicyId": cache_policy_id,
                },
                "Origins": {
                    "Quantity": 1,
                    "Items": [
                        {
                            "Id": origin_id,
                            "DomainName": self.bucket_domain_name,
                            "S3OriginConfig": {
                                "OriginAccessIdentity": "",
                            },
                        },
                    ],
                },
            }
        )

        dist = resp["Distribution"]
        self.distribution_id = dist["Id"]
        self.domain_name = dist["DomainName"]
        self.arn = dist["ARN"]

        _wait_until_deployed(
            lambda: self.cf_client.get_distribution(Id=self.distribution_id)["Distribution"],
            f"Distribution {self.name}",
        )

        logging.info(
            f"Multi-tenant Distribution bootstrapped: "
            f"id={self.distribution_id}, domain={self.domain_name}"
        )

    def cleanup(self):
        if self.distribution_id:
            logging.info(
                f"Cleaning up Distribution: {self.distribution_id}"
            )
            try:
                # Get current ETag
                resp = self.cf_client.get_distribution_config(
                    Id=self.distribution_id
                )
                etag = resp["ETag"]
                config = resp["DistributionConfig"]

                # Disable if enabled
                if config.get("Enabled", False):
                    config["Enabled"] = False
                    self.cf_client.update_distribution(
                        Id=self.distribution_id,
                        DistributionConfig=config,
                        IfMatch=etag,
                    )
                    _wait_until_deployed(
                        lambda: self.cf_client.get_distribution(
                            Id=self.distribution_id
                        )["Distribution"],
                        f"Distribution {self.distribution_id} (disable)",
                    )
                    # Re-fetch ETag after update
                    resp = self.cf_client.get_distribution_config(
                        Id=self.distribution_id
                    )
                    etag = resp["ETag"]

                self.cf_client.delete_distribution(
                    Id=self.distribution_id,
                    IfMatch=etag,
                )
            except Exception as ex:
                logging.error(
                    f"Error cleaning up Distribution "
                    f"{self.distribution_id}: {ex}"
                )

        super().cleanup()
