# ForgeOps Profile
# forgeops_profile_version=v0.0.1
#
# This file is sourced by cluster-up.sh scripts to set asset tags required
# by EntSec team.
# The profile can also be merged with some or all of the variables CDM size
# scripts eg. small.sh
# This file's name should be ~/.forgeops.env.sh
# Set the FO_ENV environment variable when running cluster-up.sh to use/create a specific profile.

# Where to get help?
# For questions around values, policy, why was my cluster deleted see #enterprise-security
# For question about  see
# For bugs and script issues #cloud-deployment

export IS_FORGEROCK=yes

# User email which might be used to contact you.
# Note:
#    - all lower case
#    - , _ for .
#    - just the username of the email
# e.g. ES_USEREMAIL=david_goldsmith
export ES_USEREMAIL=cloud_eng_team
# Obtains the priority and lifetime set by EntSec.
# production,preproduction,development,sandbox,ephemeral
export ES_ZONE=ephemeral
# UK/US/AJP
export BILLING_ENTITY=us
# Determines the party responsible for the asset.
# university,tpp,supsus,sales,sa,openbanking,marketing,fraas,it,engineering,entsec,dss,ctooffice,backstage,autoeng,am-engineering
export ES_BUSINESSUNIT=engineering
# These two can be more grainular or just the same as ES_BUSINESSUNIT
# but deviations supposed to set up with EntSec
export ES_OWNEDBY=cdm
export ES_MANAGEDBY=cdm
# Source these values for a small cluster

# Change cluster name to a unique name that can include alphanumeric characters and hyphens only.
export NAME="ds-operator-test"

# cluster-up.sh retrieves the region from the user's gcloud config.
# NODE_LOCATIONS refers to the zones to be used by CDM in the region. If your region doesn't include zones a,b or c then uncomment and set the REGION, ZONE and NODE_LOCATIONS appropriately to override:
# export REGION=us-east1
# export NODE_LOCATIONS="$REGION-b,$REGION-c,$REGION-d"
# export ZONE="$REGION-b" # required for cluster master

# PRIMARY NODE POOL VALUES
export MACHINE=e2-standard-8

# DS NODE POOL VALUES
export CREATE_DS_POOL=false
export DS_MACHINE=n2-standard-8

# Values for creating a static IP
export CREATE_STATIC_IP=false # set to true to create a static IP.
# export STATIC_IP_NAME="" # uncomment to provide a unique name(defaults to cluster name).  Lowercase letters, numbers, hyphens allowed.
export DELETE_STATIC_IP=false # set to true to delete static IP, named above, when running cluster-down.sh

export REGION=us-west2
export NETWORK="projects/$PROJECT/global/networks/forgeops"
export SUB_NETWORK="projects/$PROJECT/regions/$REGION/subnetworks/forgeops"

export RELEASE_CHANNEL="regular"
export KUBE_VERSION="1.20"
