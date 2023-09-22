# docker-bake.hcl

# Build configuration variables
variable "REGISTRY" {
  default = "us-docker.pkg.dev"
}

variable "REPOSITORY" {
  default = "forgeops-public/images"
}

variable "CACHE_REGISTRY" {
  default = REGISTRY
}

variable "CACHE_REPOSITORY" {
  default = REPOSITORY
}

variable "NO_CACHE" {
  default = false
}

variable "PULL" {
  default = true
}

variable "BUILD_ARCH" {
  default = "amd64,arm64"
}

variable "BUILD_TAG" {
  default = "dev"
}


# Helper functions
function "platforms" {
  params = [BUILD_ARCH]
  result = "${formatlist("linux/%s", "${split(",", "${BUILD_ARCH}")}")}"
}

function "tags" {
  params = [REGISTRY, REPOSITORY, image, BUILD_TAG]
  result = "${formatlist("${REGISTRY}/${REPOSITORY}/${image}:%s", "${split(",", "${BUILD_TAG}")}")}"
}

# Build targets
group "default" {
  targets = [
    "ds-operator",
  ]
}

target "base" {
  context = "."
  platforms = "${platforms("${BUILD_ARCH}")}"
  no-cache = NO_CACHE
  pull = PULL
  output = ["type=registry"]
}

target "ds-operator" {
  inherits = ["base"]

  context = "."
  dockerfile = "Dockerfile"

  tags = "${tags("${REGISTRY}", "${REPOSITORY}", "ds-operator", "${BUILD_TAG}")}"
  cache-to = ["mode=max,type=registry,ref=${CACHE_REGISTRY}/${CACHE_REPOSITORY}/ds:build-cache"]
  cache-from = ["type=registry,ref=${CACHE_REGISTRY}/${CACHE_REPOSITORY}/ds-operator:build-cache"]
}

