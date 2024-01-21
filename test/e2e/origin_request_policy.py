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

"""Utilities for working with OriginRequestPolicy resources"""

import datetime
import time

import boto3
import pytest

DEFAULT_WAIT_UNTIL_EXISTS_TIMEOUT_SECONDS = 60*10
DEFAULT_WAIT_UNTIL_EXISTS_INTERVAL_SECONDS = 15
DEFAULT_WAIT_UNTIL_DELETED_TIMEOUT_SECONDS = 60*10
DEFAULT_WAIT_UNTIL_DELETED_INTERVAL_SECONDS = 15


def wait_until_exists(
        origin_request_policy_id: str,
        timeout_seconds: int = DEFAULT_WAIT_UNTIL_EXISTS_TIMEOUT_SECONDS,
        interval_seconds: int = DEFAULT_WAIT_UNTIL_EXISTS_INTERVAL_SECONDS,
    ) -> None:
    """Waits until a OriginRequestPolicy with a supplied name is returned from
    CloudFront GetOriginRequestPolicy API.

    Usage:
        from e2e.origin_request_policy import wait_until_exists

        wait_until_exists(origin_request_policy_id)

    Raises:
        pytest.fail upon timeout
    """
    now = datetime.datetime.now()
    timeout = now + datetime.timedelta(seconds=timeout_seconds)

    while True:
        if datetime.datetime.now() >= timeout:
            pytest.fail(
                "Timed out waiting for OriginRequestPolicy to exist "
                "in CloudFront API"
            )
        time.sleep(interval_seconds)

        latest = get(origin_request_policy_id)
        if latest is not None:
            break


def wait_until_deleted(
        origin_request_policy_id: str,
        timeout_seconds: int = DEFAULT_WAIT_UNTIL_DELETED_TIMEOUT_SECONDS,
        interval_seconds: int = DEFAULT_WAIT_UNTIL_DELETED_INTERVAL_SECONDS,
    ) -> None:
    """Waits until a OriginRequestPolicy with a supplied ID is no longer returned from
    the CloudFront API.

    Usage:
        from e2e.origin_request_policy import wait_until_deleted

        wait_until_deleted(origin_request_policy_id)

    Raises:
        pytest.fail upon timeout
    """
    now = datetime.datetime.now()
    timeout = now + datetime.timedelta(seconds=timeout_seconds)

    while True:
        if datetime.datetime.now() >= timeout:
            pytest.fail(
                "Timed out waiting for OriginRequestPolicy to be "
                "deleted in CloudFront API"
            )
        time.sleep(interval_seconds)

        latest = get(origin_request_policy_id)
        if latest is None:
            break


def get(origin_request_policy_id):
    """Returns a dict containing the OriginRequestPolicy record from the CloudFront
    API.

    If no such OriginRequestPolicy exists, returns None.
    """
    c = boto3.client('cloudfront')
    try:
        resp = c.get_origin_request_policy(Id=origin_request_policy_id)
        return resp['OriginRequestPolicy']
    except c.exceptions.NoSuchOriginRequestPolicy:
        return None
