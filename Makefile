DOCKER_REPO:=$(or ${DOCKER_REPO}, dosko64/geoget)
DOCKER_TAG:=$(or ${DOCKER_TAG}, latest)

MAXMINDRC=~/.maxmindrc
MAXMIND_LICENSE_KEY:=$(or ${MAXMIND_LICENSE_KEY}, $(cat "$(MAXMINDRC)"))

URL?=https://download.maxmind.com/app/geoip_download
EDITION:=$(or ${MAXMIND_DB}, GeoLite2-City)
SUFFIX?=tar.gz
FULL_URL?=$(URL)?edition_id=$(EDITION)&suffix=$(SUFFIX)&license_key=$(MAXMIND_LICENSE_KEY)
LOCAL_ARCHIVE?=$(EDITION).$(SUFFIX)

.PHONY: release

db:
	curl --fail -L -o "$(LOCAL_ARCHIVE)" "$(FULL_URL)" && \
	tar xvfz $(LOCAL_ARCHIVE) && \
	cp ./$(EDITION)*/*.mmdb ./data/maxmind.mmdb && \
	rm -rf Geo*

build:
	go build -o geoget .

clean:
	rm -rf geoget
	rm -rf data/*
	rm -rf Geo*

init:
	go get github.com/oxequa/realize

run: 
	realize start

docker:
	docker build -t $(DOCKER_REPO):$(DOCKER_TAG) .

release: clean db docker