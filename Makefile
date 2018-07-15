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

main: controller server/main.go

server/main.go:
	@mv vendor vendor.bak
	@goagen main -d $(REPO)/design >/dev/null
	@mkdir -p server
	@mv main.go server
	@rm *.go
	@echo 'server/main.go'
	@echo '1. Change package from "main" to "server"'
	@echo '2. Add "net/http" to import section'
	@echo '3. Add "github.com/groovenauts/blocks-concurrent-batch-server/controller" to import section'
	@echo '4. Change "func main()" to "func init()"'
	@echo '5. Add "controller." before each "NewXxxxController"'
	@echo '6. Comment out the lines below the comment "Start service"'
	@echo '7. Add http.HandleFunc("/", service.Mux.ServeHTTP) at the end of init func'
	@mv vendor.bak vendor

app:
	@mv vendor vendor.bak
	@goagen app -d $(REPO)/design
	@mv vendor.bak vendor

controller: goa_controller converter

goa_controller:
	@mv vendor vendor.bak
	@mkdir -p controller
	@goagen controller  -d $(REPO)/design --pkg controller --out controller --app-pkg ../app
	@mv vendor.bak vendor

converter: converter_gen
converter_gen:
	@goa_model_gen converter design/*.yaml

model: model_gen
model_gen:
	@goa_model_gen model design/*.yaml

clean:
	@rm -rf app
	@rm -rf client
	@rm -rf tool
	@rm -rf swagger

goa_gen:
	@mv vendor vendor.bak
	@goagen app     -d $(REPO)/design
	@goagen swagger -d $(REPO)/design
	@goagen client  -d $(REPO)/design
	@mv vendor.bak vendor

generate: goa_gen model_gen converter_gen

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

fmt:
	@gofmt -w controller/*.go model/*.go

test:
	goapp test github.com/groovenauts/blocks-concurrent-batch-server/model

build:
	goapp build github.com/groovenauts/blocks-concurrent-batch-server/server

deploy:
	goapp deploy -application $(GAE_PROJECT) ./app

rollback:
	appcfg.py rollback ./app -A $(GAE_PROJECT)

local:
	dev_appserver.py --enable_console --skip_sdk_update_check=yes server/app.yaml
# goapp serve ./server
