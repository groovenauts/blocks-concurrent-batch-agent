##### Convenient command ######

REPO:=github.com/groovenauts/blocks-concurrent-batch-server
GAE_PROJECT:=projectName

init: install bootstrap import
gen: clean generate import

bootstrap:
	@goagen bootstrap -d $(REPO)/design

clean:
	@rm -rf app
	@rm -rf client
	@rm -rf tool
	@rm -rf swagger

generate:
	@goagen app     -d $(REPO)/design
	@goagen swagger -d $(REPO)/design
	@goagen client  -d $(REPO)/design

install:
	@which dep || go get -u github.com/golang/dep/cmd/dep
	@dep ensure

import:
	@which gorep || go get -v github.com/novalagung/gorep
	@gorep -path="./" \
          -from="../app" \
          -to="$(REPO)/app"
	@gorep -path="./" \
          -from="../client" \
          -to="$(REPO)/client"
	@gorep -path="./" \
          -from="../tool/cli" \
          -to="$(REPO)/tool/cli"

deploy:
	goapp deploy -application $(GAE_PROJECT) ./app

rollback:
	appcfg.py rollback ./app -A $(GAE_PROJECT)

local:
	goapp serve ./server
