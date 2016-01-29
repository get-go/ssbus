
cgo:
	@echo "Compiling with CGO enabled ...."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
	@if [[ "$?" -eq "0" ]]; then echo "Success!"; else echo "Failure!"; exit 1; fi

docker: cgo
	@echo "Building docker file ..."
	@echo "FROM scratch" > Dockerfile
	@echo "ADD main /" >> Dockerfile
	@echo "" >> Dockerfile
	@echo "EXPOSE 8080" >> Dockerfile
	@echo "" >> Dockerfile
	@echo "CMD [\"/main\"]" >> Dockerfile
	@if [[ "$?" -eq "0" ]]; then echo "Success!"; else echo "Failure!"; exit 1; fi

