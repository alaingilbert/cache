PKGS       = $(shell go list ./... | grep -v /vendor/ | grep -v /bindata)

lint:
	@golint $(PKGS)

test:
	@go test $(PKGS)

cover:
	@mkdir -p ./coverage
	@for pkg in $(PKGS) ; do \
		go test \
			-coverpkg=$$(go list -f '{{ join .Deps "\n" }}' $$pkg | grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | tr '\n' ',')$$pkg \
			-coverprofile="./coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	done
	@gocovmerge ./coverage/*.cover > cover.out
	@go tool cover -html=cover.out