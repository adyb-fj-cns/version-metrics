include .env
include staging/.env
export $(shell sed 's/=.*//' .env)
export $(shell sed 's/=.*//' staging/.env)

local: build-image push-image test-image

staging: staging-build-image staging-push-image staging-apply

staging-build-image:
	docker build -t ${STAGING_IMAGE} .

staging-push-image:
	docker push ${STAGING_IMAGE}

staging-apply:
	kubectx ${STAGING_CONTEXT}
	kubectl apply -f staging/k8s.yaml

build-image:
	docker build -t ${DEV_IMAGE} .

push-image:
	docker push ${DEV_IMAGE}

test-image:
	kubectl run ${APP} -it --rm --image=${DEV_IMAGE} --port 8000

dev:
	kubectx ${DEV_CONTEXT} 
	tilt up --stream

cleanup:
	tilt down