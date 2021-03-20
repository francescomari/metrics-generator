build:
	docker build -t francescomari/metrics-generator .

push: build
	docker push francescomari/metrics-generator
