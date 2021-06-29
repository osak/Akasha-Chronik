all: akasha-chronik

savers: build/save_pixiv build/save_twitter build/save_fanbox

build/%:
	go build -o $@ .../cmd/$*

akasha-chronik: savers docker/Dockerfile
	docker build -f docker/Dockerfile \
		-t $@ \
		build