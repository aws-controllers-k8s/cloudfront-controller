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

"""Utilities for working with VPC Block Public Access (BPA).

CloudFront refuses to deploy a VpcOrigin whose endpoint lives in a subnet
that VPC BPA applies to. When BPA is enabled for the test account and region,
the VpcOrigin tests exempt the origin's subnets for the duration of the test
by creating subnet-scoped BPA exclusions, and remove them afterwards.
"""

import datetime
import logging
import time
from typing import List

import boto3
import pytest
from botocore.exceptions import ClientError

from acktest.aws.identity import get_region

# allow-bidirectional is the only exclusion mode that has any effect when the
# regional BPA mode is block-ingress, and is also valid for
# block-bidirectional.
EXCLUSION_MODE = "allow-bidirectional"

DEFAULT_WAIT_TIMEOUT_SECONDS = 60 * 10
DEFAULT_WAIT_INTERVAL_SECONDS = 10

logger = logging.getLogger(__name__)


def _ec2_client():
    return boto3.client("ec2", region_name=get_region())


def is_enabled() -> bool:
    """Returns True if VPC BPA is enabled for the current account and region.

    Raises:
        pytest.fail if BPA is enabled but exclusions are not allowed (for
        example, when BPA is enforced by an organization-level policy), since
        the VpcOrigin tests cannot pass in that case.
    """
    opts = _ec2_client().describe_vpc_block_public_access_options()[
        "VpcBlockPublicAccessOptions"
    ]
    mode = opts.get("InternetGatewayBlockMode", "off")
    if mode == "off":
        return False

    exclusions_allowed = opts.get("ExclusionsAllowed", "allowed")
    if exclusions_allowed != "allowed":
        pytest.fail(
            f"VPC Block Public Access is enabled (mode={mode}, "
            f"managedBy={opts.get('ManagedBy')}) and exclusions are "
            f"{exclusions_allowed}; CloudFront VpcOrigins cannot be deployed "
            "in this account and region"
        )
    logger.info("VPC Block Public Access is enabled (mode=%s)", mode)
    return True


def get_load_balancer_subnet_ids(load_balancer_arn: str) -> List[str]:
    """Returns the IDs of the subnets an internal load balancer is attached
    to.

    Raises:
        pytest.fail if the load balancer is not internal. Exempting the
        subnets of an internet-facing load balancer from BPA would make it
        reachable from the internet, which these tests must never do.
    """
    elbv2 = boto3.client("elbv2", region_name=get_region())
    lb = elbv2.describe_load_balancers(LoadBalancerArns=[load_balancer_arn])[
        "LoadBalancers"
    ][0]
    if lb["Scheme"] != "internal":
        pytest.fail(
            f"Refusing to exempt subnets of load balancer {load_balancer_arn} "
            f"from VPC BPA: scheme is {lb['Scheme']}, expected internal"
        )
    return [az["SubnetId"] for az in lb["AvailabilityZones"]]


def create_subnet_exclusions(subnet_ids: List[str]) -> List[str]:
    """Creates a BPA exclusion for each supplied subnet and returns the
    exclusion IDs. Does not wait for the exclusions to become active. If any
    creation fails, the exclusions created so far are deleted before the
    error is re-raised.
    """
    ec2 = _ec2_client()
    exclusion_ids = []
    try:
        for subnet_id in subnet_ids:
            resp = ec2.create_vpc_block_public_access_exclusion(
                SubnetId=subnet_id,
                InternetGatewayExclusionMode=EXCLUSION_MODE,
            )
            exclusion_id = resp["VpcBlockPublicAccessExclusion"]["ExclusionId"]
            logger.info(
                "Created VPC BPA exclusion %s for subnet %s",
                exclusion_id,
                subnet_id,
            )
            exclusion_ids.append(exclusion_id)
    except Exception:
        delete_exclusions(exclusion_ids)
        raise
    return exclusion_ids


def _wait_for_state(
    exclusion_ids: List[str],
    target_state: str,
    failed_state: str,
    timeout_seconds: int,
    interval_seconds: int,
    missing_is_target: bool = False,
) -> None:
    ec2 = _ec2_client()
    timeout = datetime.datetime.now() + datetime.timedelta(seconds=timeout_seconds)

    while True:
        try:
            exclusions = ec2.describe_vpc_block_public_access_exclusions(
                ExclusionIds=exclusion_ids,
            )["VpcBlockPublicAccessExclusions"]
        except ClientError as e:
            if "NotFound" not in e.response["Error"]["Code"]:
                raise
            if missing_is_target:
                return
            # A newly created exclusion is eventually consistent and can be
            # reported as not found for a short time after creation.
            exclusions = []
        states = {e["ExclusionId"]: e["State"] for e in exclusions}
        if missing_is_target:
            for exclusion_id in exclusion_ids:
                states.setdefault(exclusion_id, target_state)

        failed = [i for i, s in states.items() if s == failed_state]
        if failed:
            pytest.fail(f"VPC BPA exclusions {failed} reached state {failed_state}")
        if all(states.get(i) == target_state for i in exclusion_ids):
            return
        if datetime.datetime.now() >= timeout:
            pytest.fail(
                f"Timed out waiting for VPC BPA exclusions to reach "
                f"{target_state}; current states: {states}"
            )
        time.sleep(interval_seconds)


def wait_until_exclusions_active(
    exclusion_ids: List[str],
    timeout_seconds: int = DEFAULT_WAIT_TIMEOUT_SECONDS,
    interval_seconds: int = DEFAULT_WAIT_INTERVAL_SECONDS,
) -> None:
    """Waits until all supplied BPA exclusions are in create-complete state.

    Raises:
        pytest.fail upon timeout or if an exclusion fails to create
    """
    _wait_for_state(
        exclusion_ids,
        target_state="create-complete",
        failed_state="create-failed",
        timeout_seconds=timeout_seconds,
        interval_seconds=interval_seconds,
    )


def delete_exclusions(
    exclusion_ids: List[str],
    timeout_seconds: int = DEFAULT_WAIT_TIMEOUT_SECONDS,
    interval_seconds: int = DEFAULT_WAIT_INTERVAL_SECONDS,
) -> None:
    """Deletes the supplied BPA exclusions and waits for the deletion to
    complete. Errors are logged rather than raised so that test teardown
    can continue.
    """
    if not exclusion_ids:
        return

    ec2 = _ec2_client()
    deleted = []
    for exclusion_id in exclusion_ids:
        try:
            ec2.delete_vpc_block_public_access_exclusion(ExclusionId=exclusion_id)
            deleted.append(exclusion_id)
            logger.info("Deleted VPC BPA exclusion %s", exclusion_id)
        except Exception:
            logger.exception("Failed to delete VPC BPA exclusion %s", exclusion_id)

    if not deleted:
        return
    try:
        _wait_for_state(
            deleted,
            target_state="delete-complete",
            failed_state="delete-failed",
            timeout_seconds=timeout_seconds,
            interval_seconds=interval_seconds,
            missing_is_target=True,
        )
    except (Exception, pytest.fail.Exception):
        logger.exception(
            "VPC BPA exclusions %s did not reach delete-complete", deleted
        )
