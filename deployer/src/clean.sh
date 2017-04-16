#!/bin/bash -ex
cd `dirname $0`

source_bucket=danstutzman-lambda-example
target_bucket=${source_bucket}resized # Do not change this. Walkthrough code assumes this name
function=CreateThumbnail

rm -rf node_modules
rm -f $function.zip

aws s3 rm --recursive s3://$source_bucket
aws s3 rm --recursive s3://$target_bucket


