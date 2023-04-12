package suncache

import pb "suncache/suncachepb"

// PeerPicker 用于根据传入的key选择相应节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter PeerGetter里面的Get()方法用于从对应group查找缓存值
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
