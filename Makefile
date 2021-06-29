all: savers

savers: build/save_pixiv build/save_twitter build/save_fanbox

build/%:
	go build -o $@ .../cmd/$*