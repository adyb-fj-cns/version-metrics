APP=version-metrics
IMAGE=adybfjcns/${APP}
TAG=0.1.0.2

ACP_IMAGE=artifactory.bx.homeoffice.gov.uk/dev/${APP}

all: build-image push-image test-image

acp: acp-build-image acp-push-image acp-apply

acp-build-image:
	docker build -t ${ACP_IMAGE}:${TAG} .

acp-push-image:
	docker push ${ACP_IMAGE}:${TAG}

acp-apply:
	kubectx acp-sandpit11
	kubectl apply -f ./k8s-acp.yaml

build-image:
	docker build -t ${IMAGE}:${TAG} .

push-image:
	docker push ${IMAGE}:${TAG}

test-image:
	kubectl run ${APP} -it --rm --image=${IMAGE}:${TAG} --port 8000

dev:
	tilt up --stream

cleanup:
	tilt down