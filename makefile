
education:
	echo "#########"
	echo "You can also just run the tools.sh script with the same target, as that is all this is doing."
	echo "That way you do not need to use make as a sub-par bash."
	echo "#########"

install-proto-tools: education
	./tools.sh install-proto-tools

build-protos: education
	./tools.sh build-protos

setup: education
	./tools.sh setup
