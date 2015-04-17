build: .FORCE
	docker build -t associate-eip .

.PHONY: .FORCE
.FORCE: