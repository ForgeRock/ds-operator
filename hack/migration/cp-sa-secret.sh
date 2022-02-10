#!/usr/bin/env bash
# Sample script to migrate secret agent generated ds secrets to the tls format used by cert-manager
# This can be used as a transition mechanism to reuse a saved secret agent secret.
# After migration, the secret agent secrets should be deprecated in favour of cert-manager certs.

D=/tmp/sa-secrets

rm -fr $D
mkdir -p $D
SRC_NAMESPACE=default

# Get the DS and platform ca secrets
# Note we get secret from the default namespace to the current namespace
kubectl -n $SRC_NAMESPACE get secret ds -o json >$D/ds-secret.json
kubectl -n $SRC_NAMESPACE get secret platform-ca -o json >$D/ca.json

# extract the base64 PEM certs
jq <$D/ca.json -r  .data.'"ca.pem"' | base64 -d > $D/ca.pem
jq <$D/ds-secret.json -r  .data.'"ssl-key-pair-private.pem"' | base64 -d > $D/ssl-private.pem
jq <$D/ds-secret.json -r  .data.'"ssl-key-pair.pem"' | base64 -d > $D/ssl-keypair.pem



cat >$D/ssl-tls.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ds-ssl-keypair
type: kubernetes.io/tls
data:
  ca.crt: $(cat $D/ca.pem | base64 -w0)
  tls.crt: $(cat $D/ssl-keypair.pem | base64 -w0)
  tls.key: $(cat $D/ssl-private.pem | base64 -w0)
EOF

echo "Generated tls.yaml in $D."
echo "apply with kubectl apply -f $D/ssl-tls.yaml"

# Uncomment if you want to do this automatically
# kubectl apply -f $D/ssl-tls.yaml

# Now create the ds master keypair
jq <$D/ds-secret.json -r  .data.'"master-key-pair-private.pem"' | base64 -d > $D/master-private.pem
jq <$D/ds-secret.json -r  .data.'"master-key-pair.pem"' | base64 -d > $D/master-keypair.pem

cat >$D/master-tls.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ds-master-keypair
type: kubernetes.io/tls
data:
  tls.crt: $(cat $D/master-keypair.pem | base64 -w0)
  tls.key: $(cat $D/master-private.pem | base64 -w0)
EOF

echo ""
echo "Generated $D/master-tls.yaml"
echo "apply with kubectl apply -f $D/master-tls.yaml"

# Uncomment if you want to do this automatically
# kubectl apply -f $D/master-tls.yaml