IMAGE := newrelic/infrastructure-bundle

.PHONY : release
release :
	docker push ${IMAGE}:${VERSION}
	docker push ${IMAGE}:latest
