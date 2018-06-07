##### Convenient command ######

REPO:=github.com/groovenauts/blocks-concurrent-batch-server
GAE_PROJECT:=projectName

init: install bootstrap import
gen: clean generate import

# goagen douments
# https://goa.design/implement/goagen/
# https://goa.design/ja/implement/goagen/

# Rename vendor during executing goagen
#	https://github.com/goadesign/goa/issues/923#issuecomment-290424097
bootstrap: generate main

main:
	@mv vendor vendor.bak
	@goagen main -d $(REPO)/design
	@echo '1. mkdir -p server'
	@echo '2. mv main.go server/'
	@echo '3. Change package from "main" to "server"'
	@echo '4. Change "func main()" to "func init()"'
	@echo '5. Comment out the lines below the comment "Start service"'
	@echo '6. Add "net/http" to import section'
	@echo '7. Add http.HandleFunc("/", service.Mux.ServeHTTP) at the end of init func'
	@mv vendor.bak vendor

app:
	@mv vendor vendor.bak
	@goagen app -d $(REPO)/design
	@mv vendor.bak vendor

server:
	@mkdir -p server
	@mv main.go server/

controller:
	@mv vendor vendor.bak
	@mkdir -p controller
	@goagen controller  -d $(REPO)/design --pkg controller --out controller --app-pkg ../app
	@mv vendor.bak vendor

clean:
	@rm -rf app
	@rm -rf client
	@rm -rf tool
	@rm -rf swagger

generate:
	@mv vendor vendor.bak
	@goagen app     -d $(REPO)/design
	@goagen swagger -d $(REPO)/design
	@goagen client  -d $(REPO)/design
	@mv vendor.bak vendor

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

build:
	goapp build github.com/groovenauts/blocks-concurrent-batch-server/server

deploy:
	goapp deploy -application $(GAE_PROJECT) ./app

rollback:
	appcfg.py rollback ./app -A $(GAE_PROJECT)

local:
	goapp serve ./server
