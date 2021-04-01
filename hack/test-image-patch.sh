#!/usr/bin/env bash

# Test patching the image at runtime. This  should roll the statefulset
kubectl patch directoryservice/ds --type='json' \
   -p='[{"op": "replace", "path": "/spec/image", "value":"gcr.io/forgeops-public/ds-idrepo:2020.10.28-AlSugoDiNoci"}]'

