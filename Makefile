associate-eip: .FORCE
	docker build -t associate-eip .
	docker run --rm associate-eip cat /go/bin/associate-eip > associate-eip
	chmod u+x associate-eip

.PHONY: .FORCE
.FORCE: