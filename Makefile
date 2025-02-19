PROJECT_NAME := grafana-dashboards
WORKSPACE := ./tmp

.PHONY: build
build: clean
	@mkdir -p $(WORKSPACE)
	@go build -o bin/dashboards
	@zip $(WORKSPACE)/$(PROJECT_NAME).alfredworkflow info.plist bin/dashboards icons/* icon.png

.PHONY:  clean
clean:
	@rm -rf $(WORKSPACE)/*

.PHONY: install
install: build
	@open $(WORKSPACE)/$(PROJECT_NAME).alfredworkflow
