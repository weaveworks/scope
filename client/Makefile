.PHONY: run builddep build
run:
	npm start

builddep:
	npm install imagemin-gifsicle imagemin-jpegtran imagemin-optipng imagemin-pngquant

build:
	rm -rf ./dist/
	npm install
	gulp build
