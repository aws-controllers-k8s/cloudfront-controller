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
"""Bootstraps the resources required to run the CloudFront integration tests.
"""
import logging

from acktest.bootstrapping import Resources, BootstrapFailureException
from acktest.bootstrapping.s3 import Bucket
from acktest.bootstrapping.elbv2 import NetworkLoadBalancer

from e2e import bootstrap_directory
from e2e.bootstrap_resources import BootstrapResources

public_bucket_policy = """{
    "Version":"2008-10-17",
    "Statement":[{
    "Sid":"AllowPublicRead",
    "Effect":"Allow",
    "Principal": {
      "Service": "cloudfront.amazonaws.com"
    },
    "Action":["s3:GetObject"],
    "Resource":["arn:aws:s3:::$NAME/*"]
}]}"""

def service_bootstrap() -> Resources:
    logging.getLogger().setLevel(logging.INFO)

    resources = BootstrapResources(
        PublicBucket=Bucket(
            "ack-cloudfront-controller-tests",
            policy=public_bucket_policy,
        ),
        NetworkLoadBalancer=NetworkLoadBalancer(
            name_prefix="ack-cloudfront-tests",
            scheme="internal",
            num_public_subnet=1,
            num_private_subnet=1,
            apply_security_group=True
        )
    )

    try:
        resources.bootstrap()
    except BootstrapFailureException as ex:
        exit(254)

    return resources

if __name__ == "__main__":
    config = service_bootstrap()
    # Write config to current directory by default
    config.serialize(bootstrap_directory)
