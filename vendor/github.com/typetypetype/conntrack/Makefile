.PHONY: all test vet lint generate

LIBCONNTRACK=~/src/libnetfilter_conntrack
LIBNETLINK=~/src/libnfnetlink

CONNTRACK_ENUMS=nf_conntrack_query \
				nf_conntrack_attr_grp \
				ctattr_protoinfo \
				ctattr_protoinfo_tcp \
				ctattr_type \
				ctattr_tuple \
				ctattr_ip \
				ctattr_l4proto \

LINUX_ENUMS=cntl_msg_types \

all: test vet

test:
	go test

vet:
	go vet ./...

lint:
	golint .

generate:
	enum2go -package conntrack -I ${LIBNETLINK}/include/ -I ${LIBCONNTRACK}/include/ -h ${LIBCONNTRACK}/include/libnetfilter_conntrack/libnetfilter_conntrack.h ${CONNTRACK_ENUMS} > headers.go
	enum2go -package conntrack -I ${LIBNETLINK}/include/ -I ${LIBCONNTRACK}/include/ -h ${LIBCONNTRACK}/include/libnetfilter_conntrack/linux_nfnetlink_conntrack.h ${LINUX_ENUMS} > linux_headers.go
