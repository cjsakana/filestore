package main

import (
	"filestore-serve/store/ceph"
)

func main() {
	client := ceph.GetCephConn()
	ceph.CreateBucket(client, "userfile")
	ceph.ListBuckets(client)
	ceph.ListFile(client, "userfile")
}
