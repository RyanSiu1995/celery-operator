#!/bin/bash -e

operator-sdk build ryansiu1995/celery-operator:ci
docker push ryansiu1995/celery-operator:ci
operator-sdk test local ./test/e2e/
