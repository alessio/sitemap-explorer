all: build

build: sitemap-explorer

sitemap-explorer: main.go
	go build -o $@ .

clean:
	rm -f sitemap-explorer
