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

"""Utilities for working with ConnectionGroup resources"""

import datetime
import time

import boto3
import pytest

DEFAULT_WAIT_UNTIL_EXISTS_TIMEOUT_SECONDS = 60 * 10
DEFAULT_WAIT_UNTIL_EXISTS_INTERVAL_SECONDS = 15
DEFAULT_WAIT_UNTIL_DELETED_TIMEOUT_SECONDS = 60 * 10
DEFAULT_WAIT_UNTIL_DELETED_INTERVAL_SECONDS = 15


def wait_until_exists(
    connection_group_id: str,
    timeout_seconds: int = DEFAULT_WAIT_UNTIL_EXISTS_TIMEOUT_SECONDS,
    interval_seconds: int = DEFAULT_WAIT_UNTIL_EXISTS_INTERVAL_SECONDS,
) -> None:
    """Waits until a ConnectionGroup with a supplied ID is returned from
    CloudFront GetConnectionGroup API.

    Usage:
        from e2e.connection_group import wait_until_exists

        wait_until_exists(connection_group_id)

    Raises:
        pytest.fail upon timeout
    """
    now = datetime.datetime.now()
    timeout = now + datetime.timedelta(seconds=timeout_seconds)

    while True:
        if datetime.datetime.now() >= timeout:
            pytest.fail(
                "Timed out waiting for ConnectionGroup to exist "
                "in CloudFront API"
            )
        time.sleep(interval_seconds)

        latest = get(connection_group_id)
        if latest is not None:
            break


def wait_until_deleted(
    connection_group_id: str,
    timeout_seconds: int = DEFAULT_WAIT_UNTIL_DELETED_TIMEOUT_SECONDS,
    interval_seconds: int = DEFAULT_WAIT_UNTIL_DELETED_INTERVAL_SECONDS,
) -> None:
    """Waits until a ConnectionGroup with a supplied ID is no longer returned
    from the CloudFront API.

    Usage:
        from e2e.connection_group import wait_until_deleted

        wait_until_deleted(connection_group_id)

    Raises:
        pytest.fail upon timeout
    """
    now = datetime.datetime.now()
    timeout = now + datetime.timedelta(seconds=timeout_seconds)

    while True:
        if datetime.datetime.now() >= timeout:
            pytest.fail(
                "Timed out waiting for ConnectionGroup to be "
                "deleted in CloudFront API"
            )
        time.sleep(interval_seconds)

        latest = get(connection_group_id)
        if latest is None:
            break


def get(connection_group_id):
    """Returns a dict containing the ConnectionGroup record from the CloudFront
    API.

    The identifier accepts the ID, name, or ARN of the connection group.

    If no such ConnectionGroup exists, returns None.
    """
    c = boto3.client("cloudfront")
    try:
        resp = c.get_connection_group(Identifier=connection_group_id)
        return resp["ConnectionGroup"]
    except c.exceptions.EntityNotFound:
        return None

def get_tags(connection_group_arn):
    """Returns the tags for a ConnectionGroup.

    CloudFront does not return tags inline on GetConnectionGroup, so they must
    be retrieved separately via ListTagsForResource.

    If no such ConnectionGroup exists, returns None.
    """
    c = boto3.client("cloudfront")
    try:
        resp = c.list_tags_for_resource(Resource=connection_group_arn)
        return resp["Tags"]["Items"]
    except c.exceptions.NoSuchResource:
        return None
